package config

import (
	"embed"
	"net/http"
	"os"
	"time"

	"quizon/internal/app/delivery/api"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const port string = `:8000`

func NewServer(serverI api.ServerInterface, swaggerFiles embed.FS, openapiFiles embed.FS) *http.Server {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(
		cors.Handler(
			cors.Options{
				AllowedOrigins: []string{
					"localhost:3000",
					"http://localhost:3000",
					"localhost:8000",
					"http://localhost:8000",
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

	if os.Getenv("DOMAIN") == "localhost" {
		// add swagger
		router.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.FS(swaggerFiles))))
		// add openapi
		router.Handle("/openapi/*", http.StripPrefix("/openapi/", http.FileServer(http.FS(openapiFiles))))
	}

	handler := api.HandlerFromMux(serverI, router)

	server := http.Server{
		Addr:              port,
		Handler:           handler,
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 5,
		IdleTimeout:       time.Second * 5,
	}

	return &server
}
