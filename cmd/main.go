package main

import (
	"context"
	"net/http"

	"quizon/internal/app/repository"
	"quizon/internal/app/usecase"
	"quizon/internal/app/web"
	"quizon/internal/config"
	"quizon/internal/pkg/cache"
	"quizon/internal/pkg/logger"
)

func main() {
	logger.Infof("runtime start")
	ctx := context.Background()

	db, err := config.ConnectToPostgres(ctx)
	if err != nil {
		logger.Fatalf("can't start postgres: %v", err)
	}
	logger.Infof("db ready")

	cookieCache := cache.New[*http.Cookie]()

	repository := repository.NewRepository(db)
	usecase := usecase.NewUsecase(repository, cookieCache)

	webUI, err := web.New(usecase, cookieCache)
	if err != nil {
		logger.Fatalf("can't init web ui: %v", err)
	}

	server := config.NewServer(webUI.Routes)

	logger.Infof("starting server on port: %v", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatalf("can't start server: %v", err)
	}
}
