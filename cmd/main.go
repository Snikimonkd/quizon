package main

import (
	"context"

	httpDelivery "quizon/internal/app/delivery"
	"quizon/internal/app/repository"
	"quizon/internal/app/usecase"
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

	cookieCache := cache.New[usecase.Cookie]()

	repository := repository.NewRepository(db)
	usecase := usecase.NewUsecase(repository, cookieCache)
	delivery := httpDelivery.NewDelivery(usecase)

	authMiddleware := httpDelivery.NewCheckCookieMiddleware(cookieCache)
	server := config.NewServer(delivery, authMiddleware.CheckCookie)

	logger.Infof("starting server on port: %v", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatalf("can't start server: %v", err)
	}
}
