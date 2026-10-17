package main

import (
	"context"
	"log/slog"
	"net/http"

	"quizon/internal/app/repository"
	"quizon/internal/app/usecase"
	"quizon/internal/app/web"
	"quizon/internal/config"
	"quizon/internal/pkg/cache"
	"quizon/internal/pkg/logger"
)

func main() {
	debug := config.Debug()
	logger.Setup(debug)
	slog.Info("runtime start", "debug", debug)

	ctx := context.Background()

	db, err := config.ConnectToPostgres(ctx)
	if err != nil {
		logger.Fatal("can't start postgres", err)
	}
	slog.Info("db ready")

	cookieCache := cache.New[*http.Cookie]()

	repository := repository.NewRepository(db)
	usecase := usecase.NewUsecase(repository, cookieCache)

	webUI, err := web.New(usecase)
	if err != nil {
		logger.Fatal("can't init web ui", err)
	}

	server := config.NewServer(webUI.Routes, debug)

	slog.Info("starting server", "addr", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal("can't start server", err)
	}
}
