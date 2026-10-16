package web

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/pkg/logger"
	"quizon/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"
)

//go:embed templates static
var files embed.FS

// Usecase - бизнес-логика, которую использует веб-интерфейс
type Usecase interface {
	ListGames(ctx context.Context, page int64, perPage int64) ([]usecase.GameListItem, error)
	GetGame(ctx context.Context, gameID int64) (model.Games, error)
	CreateRegistration(ctx context.Context, in model.Registrations) (usecase.RegistrationStatus, error)
	CreateGame(ctx context.Context, in model.Games) (int64, error)
	ListRegistrations(ctx context.Context, gameID int64) ([]usecase.Registration, error)
	Login(ctx context.Context, login, password string) (*http.Cookie, error)
}

// CookieCache - хранилище сессий админов
type CookieCache interface {
	Get(key string) (*http.Cookie, bool)
	Del(key string)
}

// Web - html-интерфейс сервиса
type Web struct {
	usecase Usecase
	cache   CookieCache
	pages   map[string]*template.Template
	static  fs.FS
	assets  assets
	now     func() time.Time
}

// New - конструктор для Web
func New(uc Usecase, cache CookieCache) (*Web, error) {
	static, err := fs.Sub(files, "static")
	if err != nil {
		return nil, fmt.Errorf("can't open static files: %w", err)
	}

	assets, err := hashAssets(static)
	if err != nil {
		return nil, err
	}

	pages, err := parsePages(assets)
	if err != nil {
		return nil, err
	}

	return &Web{
		usecase: uc,
		cache:   cache,
		pages:   pages,
		static:  static,
		assets:  assets,
		now:     time.Now,
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
	})
}

func parsePages(assets assets) (map[string]*template.Template, error) {
	base, err := template.New("").Funcs(funcs(assets)).ParseFS(files, "templates/layout.html", "templates/partials.html")
	if err != nil {
		return nil, fmt.Errorf("can't parse base templates: %w", err)
	}

	names, err := fs.Glob(files, "templates/pages/*.html")
	if err != nil {
		return nil, fmt.Errorf("can't list page templates: %w", err)
	}

	pages := make(map[string]*template.Template, len(names))
	for _, name := range names {
		page, err := base.Clone()
		if err != nil {
			return nil, fmt.Errorf("can't clone base templates: %w", err)
		}

		page, err = page.ParseFS(files, name)
		if err != nil {
			return nil, fmt.Errorf("can't parse %s: %w", name, err)
		}

		pages[strings.TrimSuffix(path.Base(name), ".html")] = page
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
)

func funcs(assets assets) template.FuncMap {
	return template.FuncMap{
		// /static/app.css?v=<хэш>, см. assets.handler
		"asset": assets.url,
		// 2 ноября
		"ruDate": func(t time.Time) string {
			t = t.In(utils.LocMsk)
			return fmt.Sprintf("%d %s", t.Day(), months[t.Month()-1])
		},
		// суббота
		"ruWeekday": func(t time.Time) string {
			return weekdays[t.In(utils.LocMsk).Weekday()]
		},
		// сб, 2 ноября, 18:00
		"ruDateTime": func(t time.Time) string {
			t = t.In(utils.LocMsk)
			return fmt.Sprintf("%s, %d %s, %s", shortWeekdays[t.Weekday()], t.Day(), months[t.Month()-1], t.Format("15:04"))
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
		"squad": squad,
	}
}

// layoutData - общие данные для всех страниц
type layoutData struct {
	Title string
	Admin bool
	Year  int
	Page  any
}

func (wb *Web) layout(r *http.Request, title string, page any) layoutData {
	return layoutData{
		Title: title,
		Admin: wb.isAdmin(r),
		Year:  wb.now().In(utils.LocMsk).Year(),
		Page:  page,
	}
}

// renderPage - рендерит страницу целиком
func (wb *Web) renderPage(w http.ResponseWriter, status int, name string, data layoutData) {
	wb.execute(w, status, name, "layout", data)
}

// renderBlock - рендерит один блок страницы, для ответов на htmx-запросы
func (wb *Web) renderBlock(w http.ResponseWriter, status int, name, block string, data any) {
	wb.execute(w, status, name, block, data)
}

func (wb *Web) execute(w http.ResponseWriter, status int, name, block string, data any) {
	page, ok := wb.pages[name]
	if !ok {
		logger.Errorf("unknown page template: %s", name)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	err := page.ExecuteTemplate(&buf, block, data)
	if err != nil {
		logger.Errorf("can't render %s/%s: %v", name, block, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func (wb *Web) serverError(w http.ResponseWriter, err error) {
	logger.Errorf("web: %v", err)
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
