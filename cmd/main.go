package main

import (
	"context"
	"net/http"

	"quizon/internal/app/delivery"
	"quizon/internal/app/delivery/api"
	"quizon/internal/app/repository"
	"quizon/internal/app/usecase"
	"quizon/internal/config"
	"quizon/internal/pkg/cache"
	"quizon/internal/pkg/logger"
	"quizon/openapi"

	swaggerui "quizon/swagger-ui"
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

	strictServer := delivery.New(usecase)
	authMiddleware := delivery.NewCheckCookieMiddleware(cookieCache)

	strictHandler := api.NewStrictHandler(
		&strictServer,
		[]api.StrictMiddlewareFunc{delivery.LogErrors, authMiddleware.CheckCookie},
	)

	server := config.NewServer(strictHandler, swaggerui.SwaggerContent, openapi.OpenapiContent)

	logger.Infof("starting server on port: %v", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatalf("can't start server: %v", err)
	}
}
