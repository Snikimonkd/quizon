package main

import (
	"context"
	"net/http"
	"time"

	httpDelivery "quizon/internal/app/delivery/http"
	"quizon/internal/app/repository"
	"quizon/internal/app/usecase"
	"quizon/internal/config"
	"quizon/internal/pkg/logger"
)

const port string = "8080"

func main() {
	logger.Infof("runtime start")
	ctx := context.Background()

	db, err := config.ConnectToPostgres(ctx)
	if err != nil {
		logger.Fatalf("can't start postgres: %v", err)
	}
	logger.Infof("db ready")

	router := config.NewRouter()

	repository := repository.NewRepository(db)
	usecase := usecase.NewUsecase(repository)
	httpDelivery := httpDelivery.NewDelivery(usecase)

	router.Get("/games", httpDelivery.Games)
	router.Post("/register", httpDelivery.Register)
	router.Get("/register-available", httpDelivery.RegisterAvailable)
	router.Post("/login", httpDelivery.Login)

	router.Post("/create-game", httpDelivery.CreateGame)
	router.Post("/registrations", httpDelivery.Registrations)

	logger.Infof("starting server on port: %v", port)
	server := http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 5,
		IdleTimeout:       time.Second * 5,
	}
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatalf("can't start server: %v", err)
	}
}
