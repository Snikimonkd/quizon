package main

import (
	"context"
	"log/slog"

	"github.com/benbjohnson/clock"

	"quizon/internal/app/delivery"
	"quizon/internal/app/repository"
	"quizon/internal/app/usecase"
	"quizon/internal/config"
	"quizon/internal/pkg/cache"
	"quizon/internal/pkg/logger"
)

func main() {
	// формат логов задаёт конфиг, поэтому его ошибку пишет логгер по умолчанию
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("can't read config", err)
	}

	logger.Setup(cfg.Debug)
	slog.Info("runtime start", "debug", cfg.Debug)

	ctx := context.Background()

	db, err := config.ConnectToPostgres(ctx, cfg.PgDSN)
	if err != nil {
		logger.Fatal("can't start postgres", err)
	}
	slog.Info("db ready")

	sessions := cache.New[usecase.Session]()

	repository := repository.NewRepository(db)
	usecase := usecase.NewUsecase(repository, sessions, clock.New())
	delivery, err := delivery.New(usecase, cfg.Domain)
	if err != nil {
		logger.Fatal("can't init delivery ui", err)
	}

	server := config.NewServer(delivery, cfg.Debug)

	slog.Info("starting server", "addr", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal("can't start server", err)
	}
}
