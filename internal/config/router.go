package config

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

func NewRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(
		cors.Handler(
			cors.Options{
				AllowedOrigins: []string{
					"localhost:3000",
					"http://localhost:3000",
					"https://quiz-on.ru",
					"https://www.quiz-on.ru",
				},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{
					"Content-Type",
					"Origin",
					"User-Agent",
					"Sec-Fetch-Site",
					"Sec-Fetch-Mode",
					"Sec-Fetch-Dest",
					"Referer",
					"Access-Control-Request-Method",
					"Access-Control-Request-Headers",
					"Accept-Language",
					"Accept-Encoding",
					"Accept",
					"Access-Control-Allow-Headers",
					"Access-Control-Expose-Headers",
					"Access-Control-Allow-Origin",
					"Authorization",
					"X-Requested-With",
					"X-CSRF-Token",
				},
				AllowCredentials: true,
				MaxAge:           300,
				Debug:            true,
			},
		),
	)

	return router
}
