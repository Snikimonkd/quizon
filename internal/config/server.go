package config

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
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

// NewServer - собирает http-сервер, маршруты регистрирует routes. debug — см. Debug
func NewServer(routes func(chi.Router), debug bool) *http.Server {
	log := slog.Default()

	router := chi.NewRouter()

	// id запроса нужен журналу: по нему ошибка обработчика находится рядом со строкой самого запроса
	router.Use(middleware.RequestID)
	router.Use(requestLogger(log))
	router.Use(recoverer(log))

	routes(router)

	listen := addr
	if debug {
		listen = debugAddr
	}

	return &http.Server{
		Addr:              listen,
		Handler:           router,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 5,
		IdleTimeout:       time.Second * 5,
	}
}
