package config

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const (
	// addr - только loopback: снаружи запросы приходят через Caddy на том же сервере, напрямую порт недоступен
	addr string = `127.0.0.1:8000`
	// debugAddr - для локальной разработки все интерфейсы: сайт открывается с телефона по IP компьютера
	debugAddr string = `:8000`
)

// Debug - запуск для локальной разработки: задана переменная окружения DEBUG, значение не важно
func Debug() bool {
	_, ok := os.LookupEnv("DEBUG")

	return ok
}

// Domain - домен куки входа из переменной окружения DOMAIN: localhost локально, quiz-on.ru на проде
func Domain() (string, error) {
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		return "", errors.New("empty domain in env variable")
	}

	return domain, nil
}

// NewServer - собирает http-сервер вокруг handler. debug — см. Debug
func NewServer(handler http.Handler, debug bool) *http.Server {
	log := slog.Default()

	// id запроса нужен журналу: по нему ошибка обработчика находится рядом со строкой самого запроса.
	// recoverer внутри журнала: ответ 500 после паники тоже попадает в журнал
	handler = middleware.RequestID(requestLogger(log)(recoverer(log)(handler)))

	listen := addr
	if debug {
		listen = debugAddr
	}

	return &http.Server{
		Addr:              listen,
		Handler:           handler,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 5,
		IdleTimeout:       time.Second * 5,
	}
}
