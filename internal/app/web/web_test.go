package web

import (
	"context"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/utils"
)

const (
	adminToken   string = "admin-token"
	registerPath string = "/games/7/register"
	loginPath    string = "/login"
	newGamePath  string = "/admin/games/new"
	adminLogin   string = "admin"
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
	return f.teams, nil
}

func (f *fakeUsecase) Login(context.Context, string, string) (*http.Cookie, error) {
	if f.loginErr != nil {
		return nil, f.loginErr
	}

	return testCookie("new-token"), nil
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
		Location:             "МГТУ им. Н.Э.Баумана ауд. 345",
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
		Location:             "ауд. 345",
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
			"18 октября",
			"воскресенье",
			`href="/login"`,
		)
		assertNotContains(t, body, "/games/3/register", "/games/4/register", "Список участников")
	})

	t.Run("admin", func(t *testing.T) {
		t.Parallel()

		rec := do(t, h, request{method: http.MethodGet, path: "/", admin: true})
		assertContains(t, rec.Body.String(), `href="/admin/games/1/teams"`, "Создать игру", "Выйти")
	})
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

func validRegistration() url.Values {
	return url.Values{
		"captain_name":   {"Николай Бауман"},
		"phone":          {"8 (999) 888-77-66"},
		"telegram":       {"@quizonmsk"},
		"team_name":      {"КвизON"},
		"players_amount": {"6"},
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   usecase.RegistrationStatus
		err      error
		wantText string
	}{
		{name: "ok", status: usecase.StatusOk, wantText: "Поздравляем!"},
		{name: "reserve", status: usecase.StatusReserve, wantText: "Вы в резерве"},
		{name: "closed", status: usecase.StatusClosed, wantText: "Мест нет"},
		{name: "not opened yet", err: usecase.ErrNotOpenedYet, wantText: "Регистрация ещё не открыта"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := &fakeUsecase{game: testGame(), status: tt.status, registrationErr: tt.err}
			h := newTestRouter(t, uc)

			rec := do(t, h, request{method: http.MethodPost, path: registerPath, form: validRegistration(), htmx: true})
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d", rec.Code)
			}

			body := rec.Body.String()
			assertContains(t, body, tt.wantText, `role="status"`)
			// htmx получает только фрагмент
			assertNotContains(t, body, "<!doctype html>", "<form")

			got := uc.gotRegistration
			if got == nil {
				t.Fatal("registration wasn't sent to usecase")
			}
			if got.GameID != 7 || got.PlayersAmount != 6 || got.TeamName != "КвизON" {
				t.Errorf("unexpected request: %+v", got)
			}
			if got.GroupName != nil || got.TeamID != nil {
				t.Errorf("empty optional fields must be nil: %+v", got)
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
			`value="Николай Бауман"`,
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
		name     string
		game     func(g model.Games) model.Games
		wantText string
		wantForm bool
	}{
		{
			name:     "open",
			game:     func(g model.Games) model.Games { return g },
			wantText: "Зарегистрироваться",
			wantForm: true,
		},
		{
			name: "waiting",
			game: func(g model.Games) model.Games {
				g.RegistrationOpenTime = testNow.Add(time.Hour)
				return g
			},
			wantText: "Регистрация ещё не открыта",
		},
		{
			name: "finished",
			game: func(g model.Games) model.Games {
				g.StartTime = testNow.Add(-time.Hour)
				return g
			},
			wantText: "Игра уже прошла",
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
			assertContains(t, body, tt.wantText, "МГТУ им. Н.Э.Баумана ауд. 345")
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

	rec := do(t, h, request{method: http.MethodGet, path: "/admin/games/7/teams"})
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
	assertContains(t, rec.Body.String(), `value="Бауманская лига"`, `type="datetime-local"`)
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
		"location":               {"ауд. 345"},
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

func teamsRouter(t *testing.T) http.Handler {
	t.Helper()

	team := func(number int64, name string) usecase.Registration {
		return usecase.Registration{
			Number:        number,
			GameID:        7,
			TeamName:      name,
			Telegram:      "@team",
			CaptainName:   "Капитан",
			Phone:         "89990000001",
			PlayersAmount: 5,
		}
	}
	teams := []usecase.Registration{team(1, "Первые"), team(2, "Вторые"), team(3, "Третьи")}

	return newTestRouter(t, &fakeUsecase{game: testGame(), teams: teams})
}

func TestTeamsPage(t *testing.T) {
	t.Parallel()

	rec := do(t, teamsRouter(t), request{method: http.MethodGet, path: "/admin/games/7/teams", admin: true})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	body := rec.Body.String()
	assertContains(t, body, "Команд: 3", "Третьи", `class="is-reserve"`, "/admin/games/7/teams.xlsx")
	if strings.Count(body, `class="is-reserve"`) != 1 {
		t.Error("only the third team must be in reserve")
	}
}

func TestTeamsExport(t *testing.T) {
	t.Parallel()

	rec := do(t, teamsRouter(t), request{method: http.MethodGet, path: "/admin/games/7/teams.xlsx", admin: true})
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
	if rows[1][2] != "Первые" || rows[1][1] != "основа" || rows[3][1] != "резерв" {
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
