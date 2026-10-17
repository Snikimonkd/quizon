package web

import (
	"errors"
	"html/template"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"
	"go.uber.org/mock/gomock"

	"quizon/internal/app/usecase"
	"quizon/internal/app/web/mocks"
	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/pkg/testsupport"
	"quizon/internal/utils"
)

const (
	adminToken   string = "admin-token"
	registerPath string = "/games/7/register"
	loginPath    string = "/login"
	newGamePath  string = "/admin/games/new"
	teamsPath    string = "/admin/games/7/teams"
	adminLogin   string = "admin"
	// locationFact - место игры в фактах карточки
	locationFact string = "<dt>место</dt><dd>" + defaultGameLocation + "</dd>"
)

var testNow = time.Date(2026, time.October, 16, 12, 0, 0, 0, utils.LocMsk)

// mockUsecase - мок usecase без ожиданий: вызов, которого тест не задал, роняет тест
func mockUsecase(t *testing.T) *mocks.MockUsecase {
	t.Helper()

	return mocks.NewMockUsecase(gomock.NewController(t))
}

// expectGames - главная получит список игр games
func expectGames(uc *mocks.MockUsecase, games ...usecase.GameListItem) {
	uc.EXPECT().ListGames(gomock.Any(), int64(1), gamesPerPage).Return(games, nil)
}

// expectGame - loadGame найдёт игру game со статусом регистрации status
func expectGame(uc *mocks.MockUsecase, game model.Games, status usecase.RegistrationStatus) {
	uc.EXPECT().GetGame(gomock.Any(), game.ID).Return(usecase.GameListItem{Games: game, Status: status}, nil).AnyTimes()
}

// expectTeams - страница участников игры game со списком teams. Статус регистрации странице не нужен
func expectTeams(uc *mocks.MockUsecase, game model.Games, teams []usecase.Registration) {
	expectGame(uc, game, "")
	uc.EXPECT().ListRegistrations(gomock.Any(), game.ID).Return(teams, nil).AnyTimes()
}

// testTeams - три команды со случайными данными в порядке подачи. В testGame основа 2, так что третья в резерве
func testTeams() []usecase.Registration {
	team := func(number int64) usecase.Registration {
		reg := testsupport.FakeRegistration(7)
		reg.ID = number * 10

		return usecase.Registration{Registrations: reg, Number: number}
	}

	return []usecase.Registration{team(1), team(2), team(3)}
}

// withSecondDeleted - те же команды после удаления второй (id 20), как их отдаёт usecase:
// удалённая остаётся в списке без номера, третья поднимается в основу
func withSecondDeleted(teams []usecase.Registration) []usecase.Registration {
	res := slices.Clone(teams)
	deletedAt := testNow
	res[1].DeletedAt = &deletedAt
	res[1].Number = 0
	res[2].Number = 2

	return res
}

// teamName - название i-й команды так, как его выводит html/template
func teamName(teams []usecase.Registration, i int) string {
	return template.HTMLEscapeString(teams[i].TeamName)
}

func testCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     usecase.AuthorizationTokenName,
		Value:    value,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// newTestRouter - веб-интерфейс на моке uc. Сессия админа одна — adminToken
func newTestRouter(t *testing.T, uc *mocks.MockUsecase) http.Handler {
	t.Helper()

	uc.EXPECT().CheckSession(adminToken).Return(true).AnyTimes()

	wb, err := New(uc)
	if err != nil {
		t.Fatalf("can't create web: %v", err)
	}

	r := chi.NewRouter()
	wb.Routes(r)

	return r
}

type request struct {
	method string
	path   string
	form   url.Values
	htmx   bool
	admin  bool
}

func do(t *testing.T, h http.Handler, req request) *httptest.ResponseRecorder {
	t.Helper()

	var body *strings.Reader
	if req.form != nil {
		body = strings.NewReader(req.form.Encode())
	} else {
		body = strings.NewReader("")
	}

	r := httptest.NewRequestWithContext(t.Context(), req.method, req.path, body)
	if req.form != nil {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if req.htmx {
		r.Header.Set("Hx-Request", "true")
	}
	if req.admin {
		r.AddCookie(testCookie(adminToken))
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)

	return rec
}

func assertContains(t *testing.T, body string, want ...string) {
	t.Helper()

	for _, w := range want {
		if !strings.Contains(body, w) {
			t.Errorf("body doesn't contain %q", w)
		}
	}
}

func assertNotContains(t *testing.T, body string, notWant ...string) {
	t.Helper()

	for _, w := range notWant {
		if strings.Contains(body, w) {
			t.Errorf("body contains %q", w)
		}
	}
}

func testGame() model.Games {
	return model.Games{
		ID:                   7,
		Name:                 defaultGameName,
		Location:             defaultGameLocation,
		StartTime:            testNow.Add(72 * time.Hour),
		RegistrationOpenTime: testNow.Add(-time.Hour),
		MainAmount:           2,
		ReserveAmount:        1,
	}
}

func listItem(id int64, start, open time.Time, status usecase.RegistrationStatus) usecase.GameListItem {
	return usecase.GameListItem{
		ID:                   id,
		Name:                 defaultGameName,
		Location:             defaultGameLocation,
		StartTime:            start,
		RegistrationOpenTime: open,
		Status:               status,
	}
}

func TestHome(t *testing.T) {
	t.Parallel()

	future := testNow.Add(48 * time.Hour)
	games := []usecase.GameListItem{
		listItem(1, future, testNow.Add(-time.Hour), usecase.StatusOk),
		listItem(2, future, testNow.Add(-time.Hour), usecase.StatusReserve),
		listItem(3, future, testNow.Add(-time.Hour), usecase.StatusClosed),
		listItem(4, future, testNow.Add(time.Hour), usecase.StatusWaiting),
		// прошедшая игра уходит в архив
		listItem(5, testNow.Add(-time.Hour), testNow.Add(-48*time.Hour), usecase.StatusFinished),
	}
	// у каждого подтеста свой мок: подтесты параллельные
	home := func(t *testing.T, admin bool) *httptest.ResponseRecorder {
		t.Helper()

		uc := mockUsecase(t)
		expectGames(uc, games...)

		return do(t, newTestRouter(t, uc), request{method: http.MethodGet, path: "/", admin: admin})
	}

	t.Run("guest", func(t *testing.T) {
		t.Parallel()

		rec := home(t, false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d", rec.Code)
		}

		body := rec.Body.String()
		assertContains(t, body,
			"<!doctype html>",
			`href="/games/1/register">Регистрация открыта</a>`,
			`href="/games/2/register">Регистрация в резерв</a>`,
			`<p class="show__cta show__cta--off">Регистрация закрыта <span class="show__cta-value">мест нет</span></p>`,
			"data-countdown",
			// у ожидающей игры внизу такая же плашка, погашенная, с отсчётом до открытия
			`<p class="show__cta show__cta--off">Регистрация откроется <time class="countdown"`,
			// отсчёт идёт до открытия регистрации — через час после testNow. Словами время подписывает сервер
			// по настоящим часам, эти подписи проверяет TestRuOpensAt
			`<time class="countdown" datetime="2026-10-16T13:00:00`,
			"Игра завершена",
			// в углу вывески только дата и день недели, времени там нет
			`<span class="show__day">18</span> <span class="show__month">октября</span></time>`+
				` <span class="show__weekday">воскресенье</span></p>`,
			// время начала — в фактах на месте «с команды», рядом место
			locationFact,
			`<dt>начало игры</dt><dd>12:00</dd>`,
			// лампы вывески включает класс состояния
			`class="show show--open"`,
			`class="show show--reserve"`,
			`class="show show--closed"`,
			`class="show show--waiting"`,
			`href="/login"`,
			`<ul class="archive__list" role="list">`,
			// в архиве дата с годом
			`>16 октября 2026</time>, пятница</p>`,
		)
		assertNotContains(t, body,
			"/games/3/register", "/games/4/register", "Список участников",
			"Free", "с команды", "ticket__time", "show--finished",
		)

		// место слева, время начала справа
		if strings.Index(body, "<dt>место</dt>") > strings.Index(body, "<dt>начало игры</dt>") {
			t.Error("«место» must come before «начало игры»")
		}
		// дата вывески в разметке после заголовка: её читают вместе с названием игры
		if strings.Index(body, `<h2 class="show__title">`) > strings.Index(body, `<p class="show__date">`) {
			t.Error("the sign title must come before its date in markup")
		}

		// прошедшие игры — строками под своим заголовком, после всех предстоящих
		archive := strings.Index(body, `<h2 id="archive-title">Прошедшие игры</h2>`)
		if archive < 0 || strings.Index(body, `<li class="archive__row">`) < archive ||
			strings.LastIndex(body, `<article class="show `) > archive {
			t.Error("finished games must be listed under «Прошедшие игры» after the upcoming ones")
		}
	})

	t.Run("admin", func(t *testing.T) {
		t.Parallel()

		rec := home(t, true)
		assertContains(t, rec.Body.String(),
			`href="/admin/games/1/teams"`,
			"Создать игру",
			"Выйти",
			// у организатора состояние регистрации словами, ссылка на участников есть и у прошедшей игры
			`<p class="show__status">Регистрация открыта</p>`,
			`<a class="archive__teams" href="/admin/games/5/teams">Список участников</a>`,
		)
	})
}

func TestHomeOnlyPast(t *testing.T) {
	t.Parallel()

	uc := mockUsecase(t)
	expectGames(uc, listItem(5, testNow.Add(-time.Hour), testNow.Add(-48*time.Hour), usecase.StatusFinished))
	rec := do(t, newTestRouter(t, uc), request{method: http.MethodGet, path: "/"})
	body := rec.Body.String()
	assertContains(t, body, "Новых игр пока нет", "Прошедшие игры", `<li class="archive__row">`)
	assertNotContains(t, body, "Игр пока нет", `class="playbill"`)
}

func TestHomeOnlyUpcoming(t *testing.T) {
	t.Parallel()

	uc := mockUsecase(t)
	expectGames(uc, listItem(1, testNow.Add(48*time.Hour), testNow.Add(-time.Hour), usecase.StatusOk))
	rec := do(t, newTestRouter(t, uc), request{method: http.MethodGet, path: "/"})
	body := rec.Body.String()
	assertContains(t, body, `class="playbill"`)
	// без прошедших игр архива нет совсем, даже заголовка
	assertNotContains(t, body, "archive-title", "Прошедшие игры", "Игр пока нет")
}

func TestHomeEmpty(t *testing.T) {
	t.Parallel()

	uc := mockUsecase(t)
	expectGames(uc)
	rec := do(t, newTestRouter(t, uc), request{method: http.MethodGet, path: "/"})
	assertNotContains(t, rec.Body.String(), "archive-title")
	assertContains(t, rec.Body.String(),
		"Игр пока нет",
		`aria-label="КвизON во ВКонтакте"`,
		`aria-label="КвизON в Telegram"`,
		`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="m9.489`,
		`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="M11.944`,
	)
}

// validRegistration - анкета капитана со случайными данными, необязательные поля пустые
func validRegistration() url.Values {
	reg := testsupport.FakeRegistration(7)

	return url.Values{
		"captain_name":   {reg.CaptainName},
		"phone":          {reg.Phone},
		"telegram":       {reg.Telegram},
		"team_name":      {reg.TeamName},
		"players_amount": {strconv.FormatInt(reg.PlayersAmount, 10)},
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    usecase.RegistrationStatus
		err       error
		wantText  string
		wantClass string
	}{
		{
			name: "ok", status: usecase.StatusOk, wantText: "<h2>Поздравляем!</h2>",
			wantClass: `class="outcome outcome--ok marquee"`,
		},
		{
			name: "reserve", status: usecase.StatusReserve, wantText: "<h2>Ой! Вы в резерве</h2>",
			wantClass: `class="outcome outcome--reserve marquee"`,
		},
		{
			name: "closed", status: usecase.StatusClosed, wantText: "<h2>Извини! Мест нет</h2>",
			wantClass: `class="outcome outcome--closed"`,
		},
		{
			// отсчёт тикает каждую секунду, поэтому сам итог не живая область: role="status" только у заголовка
			name: "not opened yet", err: usecase.ErrNotOpenedYet, wantText: "<h2>Регистрация ещё не открыта</h2>",
			wantClass: `class="outcome outcome--waiting">`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			form := validRegistration()
			amount, err := strconv.ParseInt(form.Get("players_amount"), 10, 64)
			if err != nil {
				t.Fatal(err)
			}

			uc := mockUsecase(t)
			expectGame(uc, testGame(), usecase.StatusOk)
			// анкета доходит до usecase как есть, пустые необязательные поля становятся nil
			uc.EXPECT().CreateRegistration(gomock.Any(), model.Registrations{
				GameID:        7,
				TeamName:      form.Get("team_name"),
				CaptainName:   form.Get("captain_name"),
				Phone:         form.Get("phone"),
				Telegram:      form.Get("telegram"),
				PlayersAmount: amount,
			}).Return(tt.status, tt.err)
			h := newTestRouter(t, uc)

			rec := do(t, h, request{method: http.MethodPost, path: registerPath, form: form, htmx: true})
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d", rec.Code)
			}

			body := rec.Body.String()
			assertContains(t, body, tt.wantText, tt.wantClass, `role="status"`)
			// htmx получает только фрагмент; в заголовках итога эмодзи больше нет
			assertNotContains(t, body, "<!doctype html>", "<form", "🙈")
		})
	}
}

func TestRegisterValidation(t *testing.T) {
	t.Parallel()

	form := validRegistration()
	form.Set("phone", "123")
	form.Del("team_name")

	// анкета с ошибками до usecase не доходит: CreateRegistration мок не ждёт.
	// У каждого подтеста свой мок: подтесты параллельные
	newRouter := func(t *testing.T) http.Handler {
		t.Helper()

		uc := mockUsecase(t)
		expectGame(uc, testGame(), usecase.StatusOk)

		return newTestRouter(t, uc)
	}

	t.Run("htmx", func(t *testing.T) {
		t.Parallel()

		rec := do(t, newRouter(t), request{method: http.MethodPost, path: registerPath, form: form, htmx: true})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d", rec.Code)
		}

		body := rec.Body.String()
		assertContains(t, body,
			"<form",
			`aria-invalid="true"`,
			"Похоже, в номере ошибка",
			"Укажите название команды",
			`value="`+template.HTMLEscapeString(form.Get("captain_name"))+`"`,
		)
		assertNotContains(t, body, "<!doctype html>")
	})

	t.Run("no js", func(t *testing.T) {
		t.Parallel()

		rec := do(t, newRouter(t), request{method: http.MethodPost, path: registerPath, form: form})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d", rec.Code)
		}
		assertContains(t, rec.Body.String(), "<!doctype html>", "Укажите название команды")
	})
}

func TestRegistrationPageStates(t *testing.T) {
	t.Parallel()

	// мест нет, но игра ещё впереди: отправку всё равно решает usecase, см. TestRegisterFullGame
	const closedState = `class="booth__game show show--closed"`

	// лампы вывески показывают настоящий статус регистрации, как на главной. Лампы — картинка,
	// поэтому про резерв говорит и кнопка
	tests := []struct {
		name      string
		game      func(g model.Games) model.Games
		status    usecase.RegistrationStatus
		wantText  string
		wantState string
		wantForm  bool
		wantMore  []string
		notWant   []string
	}{
		{
			name:      "open",
			status:    usecase.StatusOk,
			wantText:  `<button type="submit">Зарегистрироваться</button>`,
			wantState: `class="booth__game show show--open"`,
			wantForm:  true,
			notWant:   []string{"в резерв"},
		},
		{
			name:      "reserve",
			status:    usecase.StatusReserve,
			wantText:  `<button type="submit">Зарегистрироваться в резерв</button>`,
			wantState: `class="booth__game show show--reserve"`,
			wantForm:  true,
		},
		{
			// мест нет: вместо формы сразу итог, как после отправки
			name:      "closed",
			status:    usecase.StatusClosed,
			wantText:  "<h2>Извини! Мест нет</h2>",
			wantState: closedState,
			wantMore:  []string{`<div class="outcome outcome--closed" role="status">`},
		},
		{
			name: "waiting",
			game: func(g model.Games) model.Games {
				g.RegistrationOpenTime = testNow.Add(time.Hour)
				return g
			},
			status:    usecase.StatusWaiting,
			wantText:  "Регистрация ещё не открыта",
			wantState: `class="booth__game show show--waiting"`,
			// отсчёт вне живой области: иначе экранный диктор перечитывал бы итог каждую секунду
			wantMore: []string{
				"data-countdown",
				`<div class="outcome outcome--waiting">`,
				`<div role="status"><h2>Регистрация ещё не открыта</h2></div>`,
				`<p>Она откроется <time class="countdown" datetime="2026-10-16T13:00:00`,
			},
		},
		{
			name: "finished",
			game: func(g model.Games) model.Games {
				g.StartTime = testNow.Add(-time.Hour)
				return g
			},
			status:    usecase.StatusFinished,
			wantText:  "Игра уже прошла",
			wantState: `class="booth__game show show--finished"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			game := testGame()
			if tt.game != nil {
				game = tt.game(game)
			}
			uc := mockUsecase(t)
			expectGame(uc, game, tt.status)
			h := newTestRouter(t, uc)

			rec := do(t, h, request{method: http.MethodGet, path: registerPath})
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d", rec.Code)
			}

			body := rec.Body.String()
			// форму и итог htmx меняет внутри #registration
			assertContains(
				t,
				body,
				tt.wantText,
				tt.wantState,
				locationFact,
				`<div id="registration" class="booth__form">`,
			)
			assertContains(t, body, tt.wantMore...)
			assertNotContains(t, body, tt.notWant...)
			if strings.Contains(body, `hx-post="`+registerPath+`"`) != tt.wantForm {
				t.Errorf("form presence = %v, want %v", !tt.wantForm, tt.wantForm)
			}
			if tt.wantForm {
				assertContains(t, body, `hx-target="#registration"`)
			}

			// POST в закрытую по времени регистрацию не доходит до usecase, даже если мест нет:
			// CreateRegistration мок не ждёт
			if !tt.wantForm && tt.wantState != closedState {
				do(t, h, request{method: http.MethodPost, path: registerPath, form: validRegistration()})
			}
		})
	}
}

// TestRegisterFullGame - по счётчику мест нет, но анкету всё равно проверяет usecase под блокировкой игры:
// пока капитан её заполнял, команду могли удалить и место освободилось
func TestRegisterFullGame(t *testing.T) {
	t.Parallel()

	uc := mockUsecase(t)
	expectGame(uc, testGame(), usecase.StatusClosed)
	// обе отправки ниже доходят до usecase
	uc.EXPECT().CreateRegistration(gomock.Any(), gomock.Any()).Return(usecase.StatusOk, nil).Times(2)
	h := newTestRouter(t, uc)

	rec := do(t, h, request{method: http.MethodPost, path: registerPath, form: validRegistration(), htmx: true})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	assertContains(t, rec.Body.String(), "<h2>Поздравляем!</h2>")

	// без js страница целиком: вывеска с состоянием до отправки, свет под итог включают правила :has(.outcome--ok)
	rec = do(t, h, request{method: http.MethodPost, path: registerPath, form: validRegistration()})
	assertContains(t, rec.Body.String(),
		"<!doctype html>", `class="booth__game show show--closed"`, "<h2>Поздравляем!</h2>",
	)
}

func TestUnknownGame(t *testing.T) {
	t.Parallel()

	uc := mockUsecase(t)
	uc.EXPECT().GetGame(gomock.Any(), int64(8)).Return(usecase.GameListItem{}, usecase.ErrNotFound)
	h := newTestRouter(t, uc)
	// «abc» не id игры: до usecase такой запрос не доходит
	for _, path := range []string{"/games/8/register", "/games/abc/register"} {
		rec := do(t, h, request{method: http.MethodGet, path: path})
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d", path, rec.Code)
		}
	}
}

func TestAdminRequiresLogin(t *testing.T) {
	t.Parallel()

	// ни гостю, ни форме новой игры usecase не нужен
	h := newTestRouter(t, mockUsecase(t))

	rec := do(t, h, request{method: http.MethodGet, path: teamsPath})
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != loginPath {
		t.Errorf("guest: status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
	}

	rec = do(t, h, request{method: http.MethodGet, path: newGamePath, htmx: true})
	if rec.Code != http.StatusNoContent || rec.Header().Get("Hx-Redirect") != loginPath {
		t.Errorf("htmx guest: status = %d, hx-redirect = %q", rec.Code, rec.Header().Get("Hx-Redirect"))
	}

	rec = do(t, h, request{method: http.MethodGet, path: newGamePath, admin: true})
	if rec.Code != http.StatusOK {
		t.Errorf("admin: status = %d", rec.Code)
	}
	assertContains(t, rec.Body.String(),
		// заголовок на погашенной вывеске, как у входа
		`<h1 class="sheet__sign marquee">Новая игра</h1>`,
		`value="Бауманская лига"`,
		`type="datetime-local"`,
		`name="main_amount" type="number" value="20"`,
		`name="reserve_amount" type="number" value="10"`,
	)

	// открытие регистрации идёт сразу за временем игры
	body := rec.Body.String()
	start := strings.Index(body, `name="start_time"`)
	open := strings.Index(body, `name="registration_open_time"`)
	location := strings.Index(body, `name="location"`)
	if start < 0 || open < start || location < open {
		t.Errorf("fields order: start_time=%d, registration_open_time=%d, location=%d", start, open, location)
	}
}

func loginForm(password string) url.Values {
	return url.Values{"login": {adminLogin}, "password": {password}}
}

func TestLogin(t *testing.T) {
	t.Parallel()

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()

		uc := mockUsecase(t)
		uc.EXPECT().Login(gomock.Any(), adminLogin, "nope").Return(nil, usecase.ErrWrongPassword)
		form := loginForm("nope")
		rec := do(t, newTestRouter(t, uc), request{method: http.MethodPost, path: loginPath, form: form})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d", rec.Code)
		}
		assertContains(t, rec.Body.String(), "Неверный логин или пароль", `value="admin"`)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		uc := mockUsecase(t)
		uc.EXPECT().Login(gomock.Any(), adminLogin, "secret").Return(testCookie("new-token"), nil)
		form := loginForm("secret")
		rec := do(t, newTestRouter(t, uc), request{method: http.MethodPost, path: loginPath, form: form})
		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
			t.Fatalf("status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
		}
		assertContains(t, rec.Header().Get("Set-Cookie"), "new-token")
	})

	t.Run("server error", func(t *testing.T) {
		t.Parallel()

		uc := mockUsecase(t)
		uc.EXPECT().Login(gomock.Any(), adminLogin, "secret").Return(nil, errors.New("db is down"))
		form := loginForm("secret")
		rec := do(t, newTestRouter(t, uc), request{method: http.MethodPost, path: loginPath, form: form})
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d", rec.Code)
		}
		assertNotContains(t, rec.Body.String(), "db is down")
	})
}

func TestCreateGame(t *testing.T) {
	t.Parallel()

	form := url.Values{
		"name":                   {"Бауманская лига"},
		"location":               {defaultGameLocation},
		"start_time":             {"2026-11-01T18:00"},
		"registration_open_time": {"2026-10-25T12:00"},
		"main_amount":            {"30"},
		"reserve_amount":         {"5"},
	}

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		uc := mockUsecase(t)
		// время из формы — московское
		uc.EXPECT().CreateGame(gomock.Any(), model.Games{
			Name:                 "Бауманская лига",
			Location:             defaultGameLocation,
			StartTime:            time.Date(2026, time.November, 1, 18, 0, 0, 0, utils.LocMsk),
			RegistrationOpenTime: time.Date(2026, time.October, 25, 12, 0, 0, 0, utils.LocMsk),
			MainAmount:           30,
			ReserveAmount:        5,
		}).Return(int64(1), nil)
		h := newTestRouter(t, uc)
		rec := do(t, h, request{method: http.MethodPost, path: newGamePath, form: form, admin: true})
		if rec.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("registration opens after start", func(t *testing.T) {
		t.Parallel()

		bad := url.Values{}
		maps.Copy(bad, form)
		bad.Set("registration_open_time", "2026-11-02T12:00")

		// неверная игра до usecase не доходит: CreateGame мок не ждёт
		h := newTestRouter(t, mockUsecase(t))
		rec := do(t, h, request{method: http.MethodPost, path: newGamePath, form: bad, admin: true})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d", rec.Code)
		}
		assertContains(t, rec.Body.String(), "Регистрация должна открыться до начала игры")
	})
}

// teamsRouter - веб-интерфейс, где у игры game команды teams
func teamsRouter(t *testing.T, game model.Games, teams []usecase.Registration) http.Handler {
	t.Helper()

	uc := mockUsecase(t)
	expectTeams(uc, game, teams)

	return newTestRouter(t, uc)
}

func TestTeamsPage(t *testing.T) {
	t.Parallel()

	teams := testTeams()
	rec := do(t, teamsRouter(t, testGame(), teams), request{method: http.MethodGet, path: teamsPath, admin: true})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	body := rec.Body.String()
	assertContains(t, body,
		"Команд: 3",
		// места, заданные при создании игры: в testGame основа 2, резерв 1
		`<dt>мест в основе</dt><dd>2</dd>`,
		`<dt>мест в резерве</dt><dd>1</dd>`,
		teamName(teams, 2),
		"/admin/games/7/teams.xlsx",
		// основной состав 2, третья команда первая в резерве: над ней линия
		`<tr class="is-reserve reserve-start">`,
		`<span class="visually-hidden">Дальше резерв. </span>3</td>`,
	)
	assertNotContains(t, body, "Состав")
	assertContains(t, body,
		`hx-post="/admin/games/7/teams/20/delete"`,
		`hx-confirm="Удалить команду «`+teamName(teams, 1)+`» из игры?"`,
		`aria-label="Удалить команду «`+teamName(teams, 1)+`»"`,
	)
	if strings.Count(body, "is-reserve") != 1 {
		t.Error("only the third team must be in reserve")
	}
}

func TestTeamsConfirmDialog(t *testing.T) {
	t.Parallel()

	const gone = "Команда потеряет место в игре и останется в списке серой, без номера. Вернуть её отсюда не получится."

	teams := testTeams()
	// место удалённой из основы займёт первая команда резерва — третья
	takeover := " Её место в основе займёт «" + teamName(teams, 2) + "» из резерва."
	rec := do(t, teamsRouter(t, testGame(), teams), request{method: http.MethodGet, path: teamsPath, admin: true})
	body := rec.Body.String()
	assertContains(t, body,
		`<dialog id="confirm-dialog"`,
		// окно — та же вывеска с лампами, что на главной
		`<article class="marquee">`,
		`<button value="cancel" class="confirm__cancel" autofocus>Отмена</button>`,
		// третья команда сама в резерве, на её место никто не придёт
		`data-confirm-text="`+gone+`"`,
	)
	// первые две в основе: про первую команду резерва говорится у обеих
	if got := strings.Count(body, `data-confirm-text="`+gone+takeover+`"`); got != 2 {
		t.Errorf("teams with reserve takeover = %d, want 2", got)
	}
}

func TestTeamsConfirmWithoutReserve(t *testing.T) {
	t.Parallel()

	game := testGame()
	game.MainAmount = 3

	rec := do(t, teamsRouter(t, game, testTeams()), request{method: http.MethodGet, path: teamsPath, admin: true})
	assertNotContains(t, rec.Body.String(), "из резерва")
}

func TestTeamsExport(t *testing.T) {
	t.Parallel()

	teams := testTeams()
	h := teamsRouter(t, testGame(), teams)
	rec := do(t, h, request{method: http.MethodGet, path: "/admin/games/7/teams.xlsx", admin: true})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != xlsxContentType {
		t.Errorf("content type = %q", rec.Header().Get("Content-Type"))
	}
	assertContains(t, rec.Header().Get("Content-Disposition"), "attachment", "filename*=")

	f, err := excelize.OpenReader(rec.Body)
	if err != nil {
		t.Fatalf("can't open xlsx: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	rows, err := f.GetRows("Команды")
	if err != nil {
		t.Fatalf("can't read rows: %v", err)
	}
	if len(rows) != 4 {
		t.Fatalf("rows = %d, want 4", len(rows))
	}
	// колонки «Состав» нет: команда идёт сразу за номером
	if rows[0][1] != "Команда" || rows[1][1] != teams[0].TeamName || rows[3][1] != teams[2].TeamName {
		t.Errorf("unexpected rows: %v", rows)
	}
}

func TestStaticCaching(t *testing.T) {
	t.Parallel()

	h := newTestRouter(t, mockUsecase(t))

	// страница ссылается на статику с версией
	page := do(t, h, request{method: http.MethodGet, path: loginPath}).Body.String()
	m := regexp.MustCompile(`/static/app\.css\?v=([0-9a-f]+)`).FindStringSubmatch(page)
	if m == nil {
		t.Fatalf("page has no versioned app.css link")
	}
	hash := m[1]

	tests := []struct {
		name      string
		path      string
		header    http.Header
		wantCode  int
		wantCache string
	}{
		{name: "current version", path: m[0], wantCode: http.StatusOK, wantCache: cacheImmutable},
		{
			name:      "old version",
			path:      "/static/app.css?v=000000000000",
			wantCode:  http.StatusOK,
			wantCache: cacheRevalidate,
		},
		{
			name:      "font from css",
			path:      "/static/fonts/ProductSans-Bold.ttf",
			wantCode:  http.StatusOK,
			wantCache: cacheRevalidate,
		},
		{
			name:      "revalidation",
			path:      "/static/app.css",
			header:    http.Header{"If-None-Match": {`"` + hash + `"`}},
			wantCode:  http.StatusNotModified,
			wantCache: cacheRevalidate,
		},
		{name: "unknown file", path: "/static/nope.css", wantCode: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil)
			maps.Copy(r.Header, tt.header)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, r)

			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if got := rec.Header().Get("Cache-Control"); got != tt.wantCache {
				t.Errorf("cache-control = %q, want %q", got, tt.wantCache)
			}
		})
	}
}

func TestHashAssetsChangesWithContent(t *testing.T) {
	t.Parallel()

	before, err := hashAssets(fstest.MapFS{"app.css": {Data: []byte("a { color: red }")}})
	if err != nil {
		t.Fatal(err)
	}
	after, err := hashAssets(fstest.MapFS{"app.css": {Data: []byte("a { color: blue }")}})
	if err != nil {
		t.Fatal(err)
	}

	oldURL, _ := before.url("app.css")
	newURL, _ := after.url("app.css")
	if oldURL == newURL {
		t.Errorf("url must change with content: %s", oldURL)
	}

	_, err = before.url("missing.css")
	if err == nil {
		t.Error("unknown asset must be an error")
	}
}

const deleteRegistrationPath = "/admin/games/7/teams/20/delete"

func TestDeleteRegistration(t *testing.T) {
	t.Parallel()

	game, teams := testGame(), testTeams()
	uc := mockUsecase(t)
	expectGame(uc, game, "")
	// обновлённый список запрашивается уже после удаления
	gomock.InOrder(
		uc.EXPECT().DeleteRegistration(gomock.Any(), game.ID, int64(20)).Return(nil),
		uc.EXPECT().ListRegistrations(gomock.Any(), game.ID).Return(withSecondDeleted(teams), nil),
	)

	req := request{method: http.MethodPost, path: deleteRegistrationPath, htmx: true, admin: true}
	rec := do(t, newTestRouter(t, uc), req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	body := rec.Body.String()
	// htmx получает только содержимое #teams с обновлённым списком, окно подтверждения остаётся на странице.
	// Места из настроек игры входят в него и после удаления не пропадают
	assertContains(t, body,
		"Команд: 2", `<dt>мест в основе</dt><dd>2</dd>`, `<dt>мест в резерве</dt><dd>1</dd>`,
		teamName(teams, 0), teamName(teams, 1), teamName(teams, 2),
	)
	assertNotContains(t, body, "<!doctype html>", "<h1>", "confirm-dialog")
}

func TestTeamsPageDeleted(t *testing.T) {
	t.Parallel()

	teams := withSecondDeleted(testTeams())
	rec := do(t, teamsRouter(t, testGame(), teams), request{method: http.MethodGet, path: teamsPath, admin: true})
	body := rec.Body.String()
	// вторая команда удалена: строка на месте, серая, без номера и без корзины; третья поднялась в основу
	assertContains(t, body,
		"Команд: 2",
		`<tr class="is-deleted" title="Команда удалена">`,
		`<td><span class="visually-hidden">Удалена</span></td>`,
		"<td>"+teamName(teams, 1)+"</td>",
	)
	assertNotContains(t, body, "is-reserve", `/teams/20/delete"`)
	if got := strings.Count(body, "hx-post="); got != 2 {
		t.Errorf("delete buttons = %d, want 2", got)
	}
}

func TestTeamsExportSkipsDeleted(t *testing.T) {
	t.Parallel()

	teams := withSecondDeleted(testTeams())
	h := teamsRouter(t, testGame(), teams)
	rec := do(t, h, request{method: http.MethodGet, path: "/admin/games/7/teams.xlsx", admin: true})

	f, err := excelize.OpenReader(rec.Body)
	if err != nil {
		t.Fatalf("can't open xlsx: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	rows, err := f.GetRows("Команды")
	if err != nil {
		t.Fatalf("can't read rows: %v", err)
	}
	if len(rows) != 3 || rows[1][1] != teams[0].TeamName || rows[2][0] != "2" || rows[2][1] != teams[2].TeamName {
		t.Errorf("export must contain only active teams: %v", rows)
	}
}

func TestDeleteRegistrationNoJS(t *testing.T) {
	t.Parallel()

	uc := mockUsecase(t)
	// без htmx после удаления только редирект на страницу участников, список здесь не нужен
	uc.EXPECT().DeleteRegistration(gomock.Any(), int64(7), int64(20)).Return(nil)
	rec := do(t, newTestRouter(t, uc), request{method: http.MethodPost, path: deleteRegistrationPath, admin: true})
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != teamsPath {
		t.Errorf("status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestDeleteRegistrationTwice(t *testing.T) {
	t.Parallel()

	uc := mockUsecase(t)
	// второе нажатие: команду уже удалило первое, и usecase отвечает ErrNotFound.
	// Это не ошибка: показываем актуальный список
	uc.EXPECT().DeleteRegistration(gomock.Any(), int64(7), int64(20)).Return(usecase.ErrNotFound)
	expectTeams(uc, testGame(), withSecondDeleted(testTeams()))

	req := request{method: http.MethodPost, path: deleteRegistrationPath, htmx: true, admin: true}
	rec := do(t, newTestRouter(t, uc), req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	assertContains(t, rec.Body.String(), "Команд: 2")
}

func TestDeleteRegistrationRejected(t *testing.T) {
	t.Parallel()

	// ни гость, ни запрос с кривым id до удаления не доходят: DeleteRegistration мок не ждёт
	h := newTestRouter(t, mockUsecase(t))

	rec := do(t, h, request{method: http.MethodPost, path: deleteRegistrationPath})
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != loginPath {
		t.Errorf("guest: status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
	}

	rec = do(t, h, request{method: http.MethodPost, path: "/admin/games/7/teams/abc/delete", admin: true})
	if rec.Code != http.StatusNotFound {
		t.Errorf("bad id: status = %d", rec.Code)
	}
}

func TestParsePagesChecksTemplates(t *testing.T) {
	t.Parallel()

	tmpl := func(text string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(text)} }
	pagesFS := func() fstest.MapFS {
		fsys := fstest.MapFS{
			"templates/layout.html":   tmpl(`{{define "layout"}}{{template "content" .}}{{end}}`),
			"templates/partials.html": tmpl(``),
		}
		for _, p := range allPages {
			fsys[p.templatePath()] = tmpl(`{{define "content"}}` + string(p) + `{{end}}`)
		}

		return fsys
	}

	_, err := parsePages(pagesFS(), assets{})
	if err != nil {
		t.Fatalf("all pages present: %v", err)
	}

	missing := pagesFS()
	delete(missing, pageTeams.templatePath())
	_, err = parsePages(missing, assets{})
	if err == nil {
		t.Error("missing template must be an error")
	}

	extra := pagesFS()
	extra["templates/pages/forgotten.html"] = tmpl(`{{define "content"}}{{end}}`)
	_, err = parsePages(extra, assets{})
	if err == nil || !strings.Contains(err.Error(), "forgotten.html") {
		t.Errorf("template without page constant must be an error, got %v", err)
	}
}

func TestPageTitles(t *testing.T) {
	t.Parallel()

	seen := make(map[pageTitle]page, len(allPages))
	for _, p := range allPages {
		title := p.title()
		if title == "" {
			t.Errorf("page %s has no title", p)
		}
		if other, ok := seen[title]; ok {
			t.Errorf("pages %s and %s have the same title %q", other, p, title)
		}
		seen[title] = p
	}

	// заголовок попадает во вкладку браузера
	rec := do(t, newTestRouter(t, mockUsecase(t)), request{method: http.MethodGet, path: loginPath})
	assertContains(t, rec.Body.String(), "<title>Вход — КвизON</title>")
}

func TestLoginSign(t *testing.T) {
	t.Parallel()

	// вход — погашенная вывеска с заголовком и подписью, форма под ней
	rec := do(t, newTestRouter(t, mockUsecase(t)), request{method: http.MethodGet, path: loginPath})
	body := rec.Body.String()
	assertContains(t, body, `<hgroup class="sheet__sign marquee">`, "<h1>Вход</h1>", "<p>Для организаторов игр</p>")
	if strings.Index(body, "sheet__sign") > strings.Index(body, `<form method="post" action="/login">`) {
		t.Error("sign must go before the form")
	}
}

func TestRuOpensAt(t *testing.T) {
	t.Parallel()

	// testNow — пятница, 16 октября 2026, 12:00 по Москве
	at := func(month time.Month, day, hour, minute int) time.Time {
		return time.Date(2026, month, day, hour, minute, 0, 0, utils.LocMsk)
	}
	tests := []struct {
		name string
		now  time.Time
		open time.Time
		want string
	}{
		{name: "same day", now: testNow, open: at(time.October, 16, 18, 0), want: "сегодня в 18:00"},
		{name: "next day", now: testNow, open: at(time.October, 17, 20, 0), want: "завтра в 20:00"},
		{name: "weekday", now: testNow, open: at(time.October, 18, 12, 0), want: "в воскресенье в 12:00"},
		{name: "tuesday takes во", now: testNow, open: at(time.October, 20, 12, 0), want: "во вторник в 12:00"},
		{name: "last weekday of the week", now: testNow, open: at(time.October, 22, 23, 30), want: "в четверг в 23:30"},
		// через неделю тот же день недели был бы двусмысленным: уже дата
		{name: "a week ahead", now: testNow, open: at(time.October, 23, 12, 0), want: "23 октября в 12:00"},
		{name: "next month", now: testNow, open: at(time.November, 2, 18, 0), want: "2 ноября в 18:00"},
		// за 20 минут до полуночи открытие в 00:10 — уже «завтра»
		{
			name: "after midnight", now: at(time.October, 16, 23, 50),
			open: at(time.October, 17, 0, 10), want: "завтра в 00:10",
		},
		// 21:30 UTC 16 октября — по Москве уже 00:30 17-го
		{
			name: "utc input", now: testNow,
			open: time.Date(2026, time.October, 16, 21, 30, 0, 0, time.UTC), want: "завтра в 00:30",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ruOpensAt(tt.open, tt.now); got != tt.want {
				t.Errorf("ruOpensAt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDateFuncs(t *testing.T) {
	t.Parallel()

	const (
		signDate    = `{{ruDay .}} {{ruMonth .}} {{ruWeekday .}}`
		archiveDate = `{{ruDate .}}, {{ruWeekday .}}`
	)
	tests := []struct {
		name string
		text string
		at   time.Time
		want string
	}{
		{
			// 22:30 UTC в субботу 31 октября — по Москве уже воскресенье, 1 ноября: другие и число, и месяц, и день недели
			name: "sign", text: signDate,
			at: time.Date(2026, time.October, 31, 22, 30, 0, 0, time.UTC), want: "1 ноября воскресенье",
		},
		{
			// 21:30 UTC 31 декабря — по Москве уже 1 января, и год следующий
			name: "archive after new year in Moscow", text: archiveDate,
			at: time.Date(2025, time.December, 31, 21, 30, 0, 0, time.UTC), want: "1 января 2026, четверг",
		},
		{
			name: "archive before new year in Moscow", text: archiveDate,
			at: time.Date(2025, time.December, 31, 20, 59, 0, 0, time.UTC), want: "31 декабря 2025, среда",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpl := template.Must(template.New("").Funcs(funcs(assets{})).Parse(tt.text))
			var b strings.Builder
			err := tmpl.Execute(&b, tt.at)
			if err != nil {
				t.Fatal(err)
			}
			if b.String() != tt.want {
				t.Errorf("got %q, want %q", b.String(), tt.want)
			}
		})
	}
}
