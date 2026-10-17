package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	defer db.Close()
	slog.Info("db ready")

	cookieCache := cache.New[*http.Cookie]()

	repository := repository.NewRepository(db)
	usecase := usecase.NewUsecase(repository, cookieCache)

	webUI, err := web.New(usecase, cookieCache)
	if err != nil {
		logger.Fatal("can't init web ui", err)
	}

	server := config.NewServer(webUI.Routes, debug)

	// SIGTERM присылает docker stop, SIGINT — Ctrl+C в make run. По ним сервер дорабатывает начатые запросы и выходит
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("starting server", "addr", server.Addr)
	err = config.Serve(ctx, server)
	if err != nil {
		logger.Fatal("server failed", err)
	}
	slog.Info("server stopped")
}
