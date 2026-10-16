package config

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// addr - только loopback: снаружи запросы приходят через Caddy на том же сервере, напрямую порт недоступен
const addr string = `127.0.0.1:8000`

// NewServer - собирает http-сервер, маршруты регистрирует routes
func NewServer(routes func(chi.Router)) *http.Server {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	routes(router)

	return &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 5,
		IdleTimeout:       time.Second * 5,
	}
}
