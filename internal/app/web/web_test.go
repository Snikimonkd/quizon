package web

import (
	"context"
	"errors"
	"html/template"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"

	"quizon/internal/app/usecase"
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

type fakeUsecase struct {
	mu sync.Mutex

	games           []usecase.GameListItem
	game            model.Games
	status          usecase.RegistrationStatus
	registrationErr error
	teams           []usecase.Registration
	loginErr        error

	gotRegistration *model.Registrations
	gotGame         *model.Games
	gotDeleted      []int64
}

func (f *fakeUsecase) ListGames(context.Context, int64, int64) ([]usecase.GameListItem, error) {
	return f.games, nil
}

func (f *fakeUsecase) GetGame(_ context.Context, gameID int64) (model.Games, error) {
	if gameID != f.game.ID {
		return model.Games{}, usecase.ErrNotFound
	}

	return f.game, nil
}

func (f *fakeUsecase) CreateRegistration(
	_ context.Context,
	in model.Registrations,
) (usecase.RegistrationStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gotRegistration = &in

	return f.status, f.registrationErr
}

func (f *fakeUsecase) CreateGame(_ context.Context, in model.Games) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gotGame = &in

	return 1, nil
}

func (f *fakeUsecase) ListRegistrations(context.Context, int64) ([]usecase.Registration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]usecase.Registration(nil), f.teams...), nil
}

func (f *fakeUsecase) DeleteRegistration(_ context.Context, gameID int64, registrationID int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if gameID != f.game.ID {
		return usecase.ErrNotFound
	}
	for i, team := range f.teams {
		if team.ID == registrationID && !team.Deleted() {
			deletedAt := testNow
			f.teams[i].DeletedAt = &deletedAt
			f.renumber()
			f.gotDeleted = append(f.gotDeleted, registrationID)
			return nil
		}
	}

	return usecase.ErrNotFound
}

func (f *fakeUsecase) Login(context.Context, string, string) (*http.Cookie, error) {
	if f.loginErr != nil {
		return nil, f.loginErr
	}

	return testCookie("new-token"), nil
}

// renumber - как в usecase: удалённая команда остаётся в списке без номера, остальные нумеруются по порядку
func (f *fakeUsecase) renumber() {
	var number int64
	for i := range f.teams {
		f.teams[i].Number = 0
		if !f.teams[i].Deleted() {
			number++
			f.teams[i].Number = number
		}
	}
}

// teamName - название i-й команды фикстуры так, как его выводит html/template
func (f *fakeUsecase) teamName(i int) string {
	return template.HTMLEscapeString(f.teams[i].TeamName)
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

type fakeCache struct{}

func (fakeCache) Get(key string) (*http.Cookie, bool) {
	if key != adminToken {
		return nil, false
	}

	return &http.Cookie{
		Name:     usecase.AuthorizationTokenName,
		Value:    key,
		Expires:  testNow.Add(time.Hour),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, true
}

func (fakeCache) Del(string) {}

func newTestRouter(t *testing.T, uc *fakeUsecase) http.Handler {
	t.Helper()

	wb, err := New(uc, fakeCache{})
	if err != nil {
		t.Fatalf("can't create web: %v", err)
	}
	wb.now = func() time.Time { return testNow }

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
	uc := &fakeUsecase{games: []usecase.GameListItem{
		listItem(1, future, testNow.Add(-time.Hour), usecase.StatusOk),
		listItem(2, future, testNow.Add(-time.Hour), usecase.StatusReserve),
		listItem(3, future, testNow.Add(-time.Hour), usecase.StatusClosed),
		listItem(4, future, testNow.Add(time.Hour), usecase.StatusOk),
		listItem(5, testNow.Add(-time.Hour), testNow.Add(-48*time.Hour), usecase.StatusOk),
	}}
	h := newTestRouter(t, uc)

	t.Run("guest", func(t *testing.T) {
		t.Parallel()

		rec := do(t, h, request{method: http.MethodGet, path: "/"})
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d", rec.Code)
		}

		body := rec.Body.String()
		assertContains(t, body,
			"<!doctype html>",
			`href="/games/1/register">Регистрация</a>`,
			`href="/games/2/register">Регистрация в резерв</a>`,
			"Мест нет",
			"data-countdown",
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
		)
		assertNotContains(t, body,
			"/games/3/register", "/games/4/register", "Список участников",
			"Free", "с команды", "ticket__time", "show--finished",
		)

		// прошедшие игры — строками под своим заголовком, после всех предстоящих
		archive := strings.Index(body, `<h2 id="archive-title">Прошедшие игры</h2>`)
		if archive < 0 || strings.Index(body, `<li class="archive__row">`) < archive ||
			strings.LastIndex(body, `<article class="show `) > archive {
			t.Error("finished games must be listed under «Прошедшие игры» after the upcoming ones")
		}
	})

	t.Run("admin", func(t *testing.T) {
		t.Parallel()

		rec := do(t, h, request{method: http.MethodGet, path: "/", admin: true})
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

	uc := &fakeUsecase{games: []usecase.GameListItem{
		listItem(5, testNow.Add(-time.Hour), testNow.Add(-48*time.Hour), usecase.StatusOk),
	}}
	rec := do(t, newTestRouter(t, uc), request{method: http.MethodGet, path: "/"})
	body := rec.Body.String()
	assertContains(t, body, "Новых игр пока нет", "Прошедшие игры", `<li class="archive__row">`)
	assertNotContains(t, body, "Игр пока нет", `class="playbill"`)
}

func TestHomeEmpty(t *testing.T) {
	t.Parallel()

	h := newTestRouter(t, &fakeUsecase{})
	rec := do(t, h, request{method: http.MethodGet, path: "/"})
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
			name: "ok", status: usecase.StatusOk, wantText: "Поздравляем!",
			wantClass: `class="outcome outcome--ok marquee"`,
		},
		{
			name: "reserve", status: usecase.StatusReserve, wantText: "Вы в резерве",
			wantClass: `class="outcome outcome--reserve marquee"`,
		},
		{
			name: "closed", status: usecase.StatusClosed, wantText: "Мест нет",
			wantClass: `class="outcome outcome--closed"`,
		},
		{
			name: "not opened yet", err: usecase.ErrNotOpenedYet, wantText: "Регистрация ещё не открыта",
			wantClass: `class="outcome outcome--waiting"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := &fakeUsecase{game: testGame(), status: tt.status, registrationErr: tt.err}
			h := newTestRouter(t, uc)

			form := validRegistration()
			rec := do(t, h, request{method: http.MethodPost, path: registerPath, form: form, htmx: true})
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d", rec.Code)
			}

			body := rec.Body.String()
			assertContains(t, body, tt.wantText, tt.wantClass, `role="status"`)
			// htmx получает только фрагмент
			assertNotContains(t, body, "<!doctype html>", "<form")

			got := uc.gotRegistration
			if got == nil {
				t.Fatal("registration wasn't sent to usecase")
			}
			// анкета доходит до usecase как есть, пустые необязательные поля становятся nil
			amount, err := strconv.ParseInt(form.Get("players_amount"), 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			want := model.Registrations{
				GameID:        7,
				TeamName:      form.Get("team_name"),
				CaptainName:   form.Get("captain_name"),
				Phone:         form.Get("phone"),
				Telegram:      form.Get("telegram"),
				PlayersAmount: amount,
			}
			if *got != want {
				t.Errorf("registration = %+v, want %+v", *got, want)
			}
		})
	}
}

func TestRegisterValidation(t *testing.T) {
	t.Parallel()

	uc := &fakeUsecase{game: testGame()}
	h := newTestRouter(t, uc)

	form := validRegistration()
	form.Set("phone", "123")
	form.Del("team_name")

	t.Run("htmx", func(t *testing.T) {
		t.Parallel()

		rec := do(t, h, request{method: http.MethodPost, path: registerPath, form: form, htmx: true})
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

		rec := do(t, h, request{method: http.MethodPost, path: registerPath, form: form})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d", rec.Code)
		}
		assertContains(t, rec.Body.String(), "<!doctype html>", "Укажите название команды")
	})

	// подтесты параллельные, поэтому проверяем после них
	t.Cleanup(func() {
		uc.mu.Lock()
		defer uc.mu.Unlock()
		if uc.gotRegistration != nil {
			t.Error("invalid registration was sent to usecase")
		}
	})
}

func TestRegistrationPageStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		game      func(g model.Games) model.Games
		wantText  string
		wantState string
		wantForm  bool
	}{
		{
			name:      "open",
			game:      func(g model.Games) model.Games { return g },
			wantText:  "Зарегистрироваться",
			wantState: `class="booth__game show show--open"`,
			wantForm:  true,
		},
		{
			name: "waiting",
			game: func(g model.Games) model.Games {
				g.RegistrationOpenTime = testNow.Add(time.Hour)
				return g
			},
			wantText:  "Регистрация ещё не открыта",
			wantState: `class="booth__game show show--waiting"`,
		},
		{
			name: "finished",
			game: func(g model.Games) model.Games {
				g.StartTime = testNow.Add(-time.Hour)
				return g
			},
			wantText:  "Игра уже прошла",
			wantState: `class="booth__game show show--finished"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := &fakeUsecase{game: tt.game(testGame())}
			h := newTestRouter(t, uc)

			rec := do(t, h, request{method: http.MethodGet, path: registerPath})
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d", rec.Code)
			}

			body := rec.Body.String()
			assertContains(t, body, tt.wantText, tt.wantState, locationFact)
			if strings.Contains(body, `hx-post="`+registerPath+`"`) != tt.wantForm {
				t.Errorf("form presence = %v, want %v", !tt.wantForm, tt.wantForm)
			}

			// POST в закрытую регистрацию не доходит до usecase
			if !tt.wantForm {
				do(t, h, request{method: http.MethodPost, path: registerPath, form: validRegistration()})
				if uc.gotRegistration != nil {
					t.Error("registration was sent to usecase")
				}
			}
		})
	}
}

func TestUnknownGame(t *testing.T) {
	t.Parallel()

	h := newTestRouter(t, &fakeUsecase{game: testGame()})
	for _, path := range []string{"/games/8/register", "/games/abc/register"} {
		rec := do(t, h, request{method: http.MethodGet, path: path})
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d", path, rec.Code)
		}
	}
}

func TestAdminRequiresLogin(t *testing.T) {
	t.Parallel()

	h := newTestRouter(t, &fakeUsecase{game: testGame()})

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

		h := newTestRouter(t, &fakeUsecase{loginErr: usecase.ErrWrongPassword})
		form := loginForm("nope")
		rec := do(t, h, request{method: http.MethodPost, path: loginPath, form: form})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d", rec.Code)
		}
		assertContains(t, rec.Body.String(), "Неверный логин или пароль", `value="admin"`)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		h := newTestRouter(t, &fakeUsecase{})
		form := loginForm("secret")
		rec := do(t, h, request{method: http.MethodPost, path: loginPath, form: form})
		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
			t.Fatalf("status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
		}
		assertContains(t, rec.Header().Get("Set-Cookie"), "new-token")
	})

	t.Run("server error", func(t *testing.T) {
		t.Parallel()

		h := newTestRouter(t, &fakeUsecase{loginErr: errors.New("db is down")})
		form := loginForm("secret")
		rec := do(t, h, request{method: http.MethodPost, path: loginPath, form: form})
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

		uc := &fakeUsecase{}
		h := newTestRouter(t, uc)
		rec := do(t, h, request{method: http.MethodPost, path: newGamePath, form: form, admin: true})
		if rec.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}

		got := uc.gotGame
		wantStart := time.Date(2026, time.November, 1, 18, 0, 0, 0, utils.LocMsk)
		if got == nil || !got.StartTime.Equal(wantStart) || got.MainAmount != 30 || got.ReserveAmount != 5 {
			t.Errorf("unexpected request: %+v", got)
		}
	})

	t.Run("registration opens after start", func(t *testing.T) {
		t.Parallel()

		bad := url.Values{}
		maps.Copy(bad, form)
		bad.Set("registration_open_time", "2026-11-02T12:00")

		uc := &fakeUsecase{}
		h := newTestRouter(t, uc)
		rec := do(t, h, request{method: http.MethodPost, path: newGamePath, form: bad, admin: true})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d", rec.Code)
		}
		assertContains(t, rec.Body.String(), "Регистрация должна открыться до начала игры")
		if uc.gotGame != nil {
			t.Error("invalid game was sent to usecase")
		}
	})
}

// teamsUsecase - игра с тремя командами со случайными данными: основной состав 2, третья команда в резерве
func teamsUsecase() *fakeUsecase {
	team := func(number int64) usecase.Registration {
		reg := testsupport.FakeRegistration(7)
		reg.ID = number * 10

		return usecase.Registration{Registrations: reg, Number: number}
	}
	teams := []usecase.Registration{team(1), team(2), team(3)}

	return &fakeUsecase{game: testGame(), teams: teams}
}

func TestTeamsPage(t *testing.T) {
	t.Parallel()

	uc := teamsUsecase()
	rec := do(t, newTestRouter(t, uc), request{method: http.MethodGet, path: teamsPath, admin: true})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	body := rec.Body.String()
	assertContains(t, body,
		"Команд: 3",
		uc.teamName(2),
		"/admin/games/7/teams.xlsx",
		// основной состав 2, третья команда первая в резерве: над ней линия
		`<tr class="is-reserve reserve-start">`,
		`<span class="visually-hidden">Дальше резерв. </span>3</td>`,
	)
	assertNotContains(t, body, "Состав")
	assertContains(t, body,
		`hx-post="/admin/games/7/teams/20/delete"`,
		`hx-confirm="Удалить команду «`+uc.teamName(1)+`» из игры?"`,
		`aria-label="Удалить команду «`+uc.teamName(1)+`»"`,
	)
	if strings.Count(body, "is-reserve") != 1 {
		t.Error("only the third team must be in reserve")
	}
}

func TestTeamsConfirmDialog(t *testing.T) {
	t.Parallel()

	const gone = "Команда потеряет место в игре и останется в списке серой, без номера. Вернуть её отсюда не получится."

	uc := teamsUsecase()
	// место удалённой из основы займёт первая команда резерва — третья
	takeover := " Её место в основе займёт «" + uc.teamName(2) + "» из резерва."
	rec := do(t, newTestRouter(t, uc), request{method: http.MethodGet, path: teamsPath, admin: true})
	body := rec.Body.String()
	assertContains(t, body,
		`<dialog id="confirm-dialog"`,
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

	uc := teamsUsecase()
	uc.game.MainAmount = 3

	rec := do(t, newTestRouter(t, uc), request{method: http.MethodGet, path: teamsPath, admin: true})
	assertNotContains(t, rec.Body.String(), "из резерва")
}

func TestTeamsExport(t *testing.T) {
	t.Parallel()

	uc := teamsUsecase()
	rec := do(t, newTestRouter(t, uc), request{method: http.MethodGet, path: "/admin/games/7/teams.xlsx", admin: true})
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
	if rows[1][2] != uc.teams[0].TeamName || rows[1][1] != "основа" || rows[3][1] != "резерв" {
		t.Errorf("unexpected rows: %v", rows)
	}
}

func TestStaticCaching(t *testing.T) {
	t.Parallel()

	h := newTestRouter(t, &fakeUsecase{})

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

	uc := teamsUsecase()
	req := request{method: http.MethodPost, path: deleteRegistrationPath, htmx: true, admin: true}
	rec := do(t, newTestRouter(t, uc), req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	body := rec.Body.String()
	// htmx получает только содержимое #teams с обновлённым списком, окно подтверждения остаётся на странице
	assertContains(t, body, "Команд: 2", uc.teamName(0), uc.teamName(1), uc.teamName(2))
	assertNotContains(t, body, "<!doctype html>", "<h1>", "confirm-dialog")
	if len(uc.gotDeleted) != 1 || uc.gotDeleted[0] != 20 {
		t.Errorf("deleted = %v, want [20]", uc.gotDeleted)
	}
}

func deletedRegistrationUsecase(t *testing.T) *fakeUsecase {
	t.Helper()

	uc := teamsUsecase()
	err := uc.DeleteRegistration(t.Context(), 7, 20)
	if err != nil {
		t.Fatalf("can't delete team: %v", err)
	}

	return uc
}

func TestTeamsPageDeleted(t *testing.T) {
	t.Parallel()

	uc := deletedRegistrationUsecase(t)
	rec := do(t, newTestRouter(t, uc), request{method: http.MethodGet, path: teamsPath, admin: true})
	body := rec.Body.String()
	// вторая команда удалена: строка на месте, серая, без номера и без корзины; третья поднялась в основу
	assertContains(t, body,
		"Команд: 2",
		`<tr class="is-deleted" title="Команда удалена">`,
		`<td><span class="visually-hidden">Удалена</span></td>`,
		"<td>"+uc.teamName(1)+"</td>",
	)
	assertNotContains(t, body, "is-reserve", `/teams/20/delete"`)
	if got := strings.Count(body, "hx-post="); got != 2 {
		t.Errorf("delete buttons = %d, want 2", got)
	}
}

func TestTeamsExportSkipsDeleted(t *testing.T) {
	t.Parallel()

	uc := deletedRegistrationUsecase(t)
	rec := do(t, newTestRouter(t, uc), request{method: http.MethodGet, path: "/admin/games/7/teams.xlsx", admin: true})

	f, err := excelize.OpenReader(rec.Body)
	if err != nil {
		t.Fatalf("can't open xlsx: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	rows, err := f.GetRows("Команды")
	if err != nil {
		t.Fatalf("can't read rows: %v", err)
	}
	if len(rows) != 3 || rows[1][2] != uc.teams[0].TeamName || rows[2][0] != "2" || rows[2][2] != uc.teams[2].TeamName {
		t.Errorf("export must contain only active teams: %v", rows)
	}
}

func TestDeleteRegistrationNoJS(t *testing.T) {
	t.Parallel()

	h := newTestRouter(t, teamsUsecase())
	rec := do(t, h, request{method: http.MethodPost, path: deleteRegistrationPath, admin: true})
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != teamsPath {
		t.Errorf("status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestDeleteRegistrationTwice(t *testing.T) {
	t.Parallel()

	h := newTestRouter(t, teamsUsecase())
	do(t, h, request{method: http.MethodPost, path: deleteRegistrationPath, htmx: true, admin: true})

	// второе нажатие не ошибка: показываем актуальный список
	rec := do(t, h, request{method: http.MethodPost, path: deleteRegistrationPath, htmx: true, admin: true})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	assertContains(t, rec.Body.String(), "Команд: 2")
}

func TestDeleteRegistrationRejected(t *testing.T) {
	t.Parallel()

	uc := teamsUsecase()
	h := newTestRouter(t, uc)

	rec := do(t, h, request{method: http.MethodPost, path: deleteRegistrationPath})
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != loginPath {
		t.Errorf("guest: status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
	}

	rec = do(t, h, request{method: http.MethodPost, path: "/admin/games/7/teams/abc/delete", admin: true})
	if rec.Code != http.StatusNotFound {
		t.Errorf("bad id: status = %d", rec.Code)
	}

	if len(uc.gotDeleted) != 0 {
		t.Errorf("nothing must be deleted, got %v", uc.gotDeleted)
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
	rec := do(t, newTestRouter(t, &fakeUsecase{}), request{method: http.MethodGet, path: loginPath})
	assertContains(t, rec.Body.String(), "<title>Вход — КвизON</title>")
}

func TestDateFuncs(t *testing.T) {
	t.Parallel()

	tmpl := template.Must(template.New("").Funcs(funcs(assets{})).Parse(`{{ruDay .}} {{ruMonth .}}`))

	// 22:30 UTC 17 октября — по Москве уже 18 октября
	var b strings.Builder
	err := tmpl.Execute(&b, time.Date(2026, time.October, 17, 22, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if b.String() != "18 октября" {
		t.Errorf("got %q, want «18 октября»", b.String())
	}
}
