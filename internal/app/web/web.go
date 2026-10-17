package web

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"math"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/samber/lo"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/utils"
)

//go:embed templates static
var files embed.FS

// Usecase - бизнес-логика, которую использует веб-интерфейс
//
//go:generate go tool mockgen -typed -destination=mock/usecase_mock.go -package=mock . Usecase
type Usecase interface {
	ListGames(ctx context.Context, page int64, perPage int64) ([]usecase.GameListItem, error)
	GetGame(ctx context.Context, gameID int64) (usecase.GameListItem, error)
	CreateRegistration(ctx context.Context, in model.Registrations) (usecase.RegistrationStatus, error)
	CreateGame(ctx context.Context, in model.Games) (int64, error)
	ListRegistrations(ctx context.Context, gameID int64) ([]usecase.Registration, error)
	DeleteRegistration(ctx context.Context, gameID int64, registrationID int64) error
	Login(ctx context.Context, login, password string) (usecase.Session, error)
	CheckSession(token string) bool
	Logout(token string)
}

// Web - html-интерфейс сервиса
type Web struct {
	usecase Usecase
	// domain - домен куки входа
	domain string
	pages  map[page]*template.Template
	static fs.FS
	assets assets
}

// New - конструктор для Web. domain - домен куки входа: localhost локально, quiz-on.ru на проде
func New(uc Usecase, domain string) (*Web, error) {
	static, err := fs.Sub(files, "static")
	if err != nil {
		return nil, fmt.Errorf("can't open static files: %w", err)
	}

	assets, err := hashAssets(static)
	if err != nil {
		return nil, err
	}

	pages, err := parsePages(files, assets)
	if err != nil {
		return nil, err
	}

	return &Web{
		usecase: uc,
		domain:  domain,
		pages:   pages,
		static:  static,
		assets:  assets,
	}, nil
}

// Routes - регистрирует html-страницы и статику
func (wb *Web) Routes(r chi.Router) {
	r.Handle("/static/*", http.StripPrefix("/static/", wb.assets.handler(wb.static)))

	r.Get("/", wb.home)
	r.Get("/games/{id}/register", wb.registrationPage)
	r.Post("/games/{id}/register", wb.register)
	r.Get("/login", wb.loginPage)
	r.Post("/login", wb.login)
	r.Post("/logout", wb.logout)

	r.Group(func(r chi.Router) {
		r.Use(wb.requireAdmin)
		r.Get("/admin/games/new", wb.newGamePage)
		r.Post("/admin/games/new", wb.createGame)
		r.Get("/admin/games/{id}/teams", wb.teamsPage)
		r.Get("/admin/games/{id}/teams.xlsx", wb.exportTeams)
		r.Post("/admin/games/{id}/teams/{registrationID}/delete", wb.deleteRegistration)
	})
}

func parsePages(fsys fs.FS, assets assets) (map[page]*template.Template, error) {
	base, err := template.New("").Funcs(funcs(assets)).ParseFS(fsys, "templates/layout.html", "templates/partials.html")
	if err != nil {
		return nil, fmt.Errorf("can't parse base templates: %w", err)
	}

	pages := make(map[page]*template.Template, len(allPages))
	for _, p := range allPages {
		tmpl, cerr := base.Clone()
		if cerr != nil {
			return nil, fmt.Errorf("can't clone base templates: %w", cerr)
		}

		tmpl, cerr = tmpl.ParseFS(fsys, p.templatePath())
		if cerr != nil {
			return nil, fmt.Errorf("can't parse page %s: %w", p, cerr)
		}

		pages[p] = tmpl
	}

	// шаблон без константы в allPages никогда не отрендерится - скорее всего, его забыли добавить
	names, err := fs.Glob(fsys, "templates/pages/*.html")
	if err != nil {
		return nil, fmt.Errorf("can't list page templates: %w", err)
	}
	for _, name := range names {
		if _, ok := pages[page(strings.TrimSuffix(path.Base(name), ".html"))]; !ok {
			return nil, fmt.Errorf("template %s has no page constant in allPages", name)
		}
	}

	return pages, nil
}

var (
	months = [...]string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}
	weekdays      = [...]string{"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"}
	shortWeekdays = [...]string{"вс", "пн", "вт", "ср", "чт", "пт", "сб"}
	// «откроется в субботу», «во вторник»
	onWeekdays = [...]string{
		"в воскресенье", "в понедельник", "во вторник", "в среду", "в четверг", "в пятницу", "в субботу",
	}
)

// ruOpensAt - когда откроется регистрация, словами относительно now по Москве: «сегодня в 12:00»,
// «завтра в 20:00», в пределах недели «в субботу в 12:00», дальше «25 октября в 12:00».
// Меньше чем за сутки app.js заменяет текст отсчётом
func ruOpensAt(open, now time.Time) string {
	open, now = open.In(utils.LocMsk), now.In(utils.LocMsk)
	at := " в " + open.Format("15:04")

	// разница в календарных днях: переход через полночь по Москве — уже «завтра», даже если до открытия минуты
	day := func(t time.Time) time.Time { return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, utils.LocMsk) }
	days := int(math.Round(day(open).Sub(day(now)).Hours() / 24))

	switch {
	case days <= 0:
		return "сегодня" + at
	case days == 1:
		return "завтра" + at
	case days < 7:
		return onWeekdays[open.Weekday()] + at
	default:
		return fmt.Sprintf("%d %s%s", open.Day(), months[open.Month()-1], at)
	}
}

func funcs(assets assets) template.FuncMap {
	return template.FuncMap{
		// /static/app.css?v=<хэш>, см. assets.handler
		"asset": assets.url,
		// 2 ноября 2026
		"ruDate": func(t time.Time) string {
			t = t.In(utils.LocMsk)
			return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()-1], t.Year())
		},
		// 2
		"ruDay": func(t time.Time) int {
			return t.In(utils.LocMsk).Day()
		},
		// ноября
		"ruMonth": func(t time.Time) string {
			return months[t.In(utils.LocMsk).Month()-1]
		},
		// суббота
		"ruWeekday": func(t time.Time) string {
			return weekdays[t.In(utils.LocMsk).Weekday()]
		},
		// сб, 2 ноября, 18:00
		"ruDateTime": func(t time.Time) string {
			t = t.In(utils.LocMsk)
			return fmt.Sprintf(
				"%s, %d %s, %s",
				shortWeekdays[t.Weekday()],
				t.Day(),
				months[t.Month()-1],
				t.Format("15:04"),
			)
		},
		"hhmm": func(t time.Time) string {
			return t.In(utils.LocMsk).Format("15:04")
		},
		// 02.01.2006 15:04:05
		"fullTime": func(t time.Time) string {
			return t.In(utils.LocMsk).Format("02.01.2006 15:04:05")
		},
		"iso": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"deref": lo.FromPtr[string],
	}
}

// layoutData - общие данные для всех страниц
type layoutData struct {
	Title pageTitle
	Admin bool
	Year  int
	Page  any
}

// renderPage - рендерит страницу целиком: общий каркас с заголовком страницы и её данные view
func (wb *Web) renderPage(w http.ResponseWriter, r *http.Request, status int, p page, view any) {
	data := layoutData{
		Title: p.title(),
		Admin: wb.isAdmin(r),
		Year:  time.Now().In(utils.LocMsk).Year(),
		Page:  view,
	}

	wb.execute(w, r, status, p, "layout", data)
}

// renderBlock - рендерит один блок страницы, для ответов на htmx-запросы
func (wb *Web) renderBlock(w http.ResponseWriter, r *http.Request, status int, p page, block string, data any) {
	wb.execute(w, r, status, p, block, data)
}

func (wb *Web) execute(w http.ResponseWriter, r *http.Request, status int, p page, block string, data any) {
	tmpl, ok := wb.pages[p]
	if !ok {
		slog.ErrorContext(r.Context(), "unknown page template", "page", string(p), "request_id", requestID(r))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, block, data)
	if err != nil {
		slog.ErrorContext(r.Context(), "can't render template",
			"page", string(p), "block", block, "err", err, "request_id", requestID(r))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// requestID - id запроса из журнала запросов: по нему ошибка находится рядом со строкой самого запроса
func requestID(r *http.Request) string {
	return middleware.GetReqID(r.Context())
}

func (wb *Web) serverError(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "request failed", "err", err, "request_id", requestID(r))
	http.Error(w, "Что-то пошло не так. Попробуйте ещё раз через минуту.", http.StatusInternalServerError)
}

// isHTMX - запрос пришёл от htmx, а не обычной навигацией браузера
func isHTMX(r *http.Request) bool {
	return r.Header.Get("Hx-Request") == "true"
}

// redirect - перенаправление, которое понимают и браузер, и htmx
func redirect(w http.ResponseWriter, r *http.Request, url string) {
	if isHTMX(r) {
		w.Header().Set("Hx-Redirect", url)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}
