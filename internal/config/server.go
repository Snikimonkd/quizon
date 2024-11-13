package config

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const port string = `:8080`

type Delivery interface {
	Games(http.ResponseWriter, *http.Request)
	RegisterAvailable(http.ResponseWriter, *http.Request)
	Register(http.ResponseWriter, *http.Request)
	Login(http.ResponseWriter, *http.Request)

	// with auth
	CreateGame(http.ResponseWriter, *http.Request)
	Registrations(http.ResponseWriter, *http.Request)
}

func NewServer(delivery Delivery, authMiddleware func(http.Handler) http.Handler) *http.Server {
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

	router.Get("/games", delivery.Games)
	router.Get("/register-available", delivery.RegisterAvailable)
	router.Post("/register", delivery.Register)
	router.Post("/login", delivery.Login)

	authRouter := router.With(authMiddleware)

	authRouter.Post("/create-game", delivery.CreateGame)
	authRouter.Get("/registrations", delivery.Registrations)

	server := http.Server{
		Addr:              port,
		Handler:           router,
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 5,
		IdleTimeout:       time.Second * 5,
	}

	return &server
}
