package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
)

func TestNewServerAddr(t *testing.T) {
	t.Parallel()

	if got := NewServer(http.NotFoundHandler()).Addr; got != "127.0.0.1:8000" {
		t.Errorf("addr = %q, want loopback only", got)
	}
}

// TestCrossOriginProtection - формы с чужих сайтов отклоняются, а свои формы, переходы по ссылкам и запросы
// не из браузера проходят. Отказ попадает в журнал, как любой запрос
func TestCrossOriginProtection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		// site - заголовок Sec-Fetch-Site, origin - Origin. Пустые не отправляются
		site   string
		origin string
		want   int
	}{
		{name: "own form", method: http.MethodPost, site: "same-origin", want: http.StatusOK},
		{name: "foreign form", method: http.MethodPost, site: "cross-site", want: http.StatusForbidden},
		{name: "link from foreign site", method: http.MethodGet, site: "cross-site", want: http.StatusOK},
		// старый браузер без Sec-Fetch-Site: Origin сверяется с Host
		{name: "old browser, own form", method: http.MethodPost, origin: "https://quiz-on.ru", want: http.StatusOK},
		{
			name:   "old browser, foreign form",
			method: http.MethodPost,
			origin: "https://evil.example",
			want:   http.StatusForbidden,
		},
		// curl и скрипты этих заголовков не шлют
		{name: "not a browser", method: http.MethodPost, want: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log, buf := testLogger()
			called := false
			h := withMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }), log)
			req := httptest.NewRequestWithContext(t.Context(), tt.method, "https://quiz-on.ru/logout", nil)
			if tt.site != "" {
				req.Header.Set("Sec-Fetch-Site", tt.site)
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.want || called != (tt.want == http.StatusOK) {
				t.Errorf("status = %d, handler called = %v, want %d", rec.Code, called, tt.want)
			}
			if recs := records(t, buf); len(recs) != 1 || recs[0]["status"] != float64(tt.want) {
				t.Errorf("records = %v, want one request record with status %d", recs, tt.want)
			}
		})
	}
}

// testLogger - логгер, который пишет JSON в буфер, как в продакшене
func testLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer

	return slog.New(slog.NewJSONHandler(&buf, nil)), &buf
}

func records(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	var res []map[string]any
	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		var rec map[string]any
		err := json.Unmarshal([]byte(line), &rec)
		if err != nil {
			t.Fatalf("log line is not JSON: %q: %v", line, err)
		}
		res = append(res, rec)
	}

	return res
}

func TestRequestLogger(t *testing.T) {
	t.Parallel()

	log, buf := testLogger()
	h := middleware.RequestID(requestLogger(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})))
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/games/7/register", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	recs := records(t, buf)
	if len(recs) != 1 {
		t.Fatalf("records = %d, want one per request", len(recs))
	}
	rec := recs[0]
	if rec["msg"] != "request" || rec["method"] != http.MethodPost || rec["path"] != "/games/7/register" ||
		rec["status"] != float64(http.StatusCreated) {
		t.Errorf("unexpected record: %v", rec)
	}
	if id, _ := rec["request_id"].(string); id == "" {
		t.Errorf("request_id is empty: %v", rec)
	}
	if _, ok := rec["duration"]; !ok {
		t.Errorf("duration is missing: %v", rec)
	}
}

func TestRequestLoggerDefaultStatus(t *testing.T) {
	t.Parallel()

	// обработчик ничего не записал: net/http отдаст 200, его и пишем
	log, buf := testLogger()
	h := requestLogger(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	if recs := records(t, buf); recs[0]["status"] != float64(http.StatusOK) {
		t.Errorf("status = %v, want 200", recs[0]["status"])
	}
}

func TestRecoverer(t *testing.T) {
	t.Parallel()

	log, buf := testLogger()
	h := requestLogger(log)(recoverer(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}

	// сначала паника со стеком, потом строка запроса со статусом 500
	recs := records(t, buf)
	if len(recs) != 2 {
		t.Fatalf("records = %d, want panic + request", len(recs))
	}
	stack, _ := recs[0]["stack"].(string)
	if recs[0]["msg"] != "panic" || recs[0]["panic"] != "boom" || !strings.Contains(stack, "TestRecoverer") {
		t.Errorf("unexpected panic record: %v", recs[0])
	}
	if recs[1]["status"] != float64(http.StatusInternalServerError) {
		t.Errorf("request record status = %v, want 500", recs[1]["status"])
	}
}

func TestRecovererAbortHandler(t *testing.T) {
	t.Parallel()

	log, buf := testLogger()
	h := recoverer(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	// ErrAbortHandler — штатный обрыв ответа: его не пишем и отдаём дальше net/http
	defer func() {
		rec := recover()
		err, _ := rec.(error)
		if !errors.Is(err, http.ErrAbortHandler) {
			t.Errorf("recovered %v, want http.ErrAbortHandler", rec)
		}
		if buf.Len() != 0 {
			t.Errorf("abort must not be logged: %s", buf.String())
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
}
