package main

import (
	"context"
	"net/http"
	"time"

	httpDelivery "quizon_bot/internal/app/delivery/http"
	"quizon_bot/internal/app/repository"
	"quizon_bot/internal/app/usecase"
	"quizon_bot/internal/config"
	"quizon_bot/internal/pkg/logger"
)

const port string = "8080"

func main() {
	logger.Infof("runtime start")
	ctx := context.Background()

	router := config.NewRouter()

	db, err := config.ConnectToPostgres(ctx)
	if err != nil {
		logger.Fatalf("can't start postgres: %v", err)
	}
	logger.Infof("db ready")

	repository := repository.NewRepository(db)
	usecase := usecase.NewUsecase(repository)
	httpDelivery := httpDelivery.NewDelivery(usecase)

	router.Post("/register", httpDelivery.Register)
	router.Post("/registrations", httpDelivery.Registrations)
	router.Get("/register-available", httpDelivery.RegisterAvailable)
	//	router.Post("/login", httpDelivery.Login)

	logger.Infof("starting server on port: %v", port)
	server := http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadTimeout:       time.Second,
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		IdleTimeout:       time.Second,
	}
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatalf("can't start server: %v", err)
	}
}
