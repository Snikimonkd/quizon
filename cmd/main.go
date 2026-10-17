package main

import (
	"context"
	"log/slog"

	"github.com/benbjohnson/clock"

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

	domain, err := config.Domain()
	if err != nil {
		logger.Fatal("can't read config", err)
	}

	ctx := context.Background()

	db, err := config.ConnectToPostgres(ctx)
	if err != nil {
		logger.Fatal("can't start postgres", err)
	}
	slog.Info("db ready")

	sessions := cache.New[usecase.Session]()

	repository := repository.NewRepository(db)
	usecase := usecase.NewUsecase(repository, sessions, clock.New())

	webUI, err := web.New(usecase, domain)
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
