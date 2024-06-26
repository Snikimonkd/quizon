package main

import (
	"context"
	"net/http"

	"github.com/Snikimonkd/quizon/internal/app/delivery"
	"github.com/Snikimonkd/quizon/internal/app/repository"
	"github.com/Snikimonkd/quizon/internal/app/usecase"
	"github.com/Snikimonkd/quizon/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func main() {
	ctx := context.Background()
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	err := config.ReadConfigFile()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	postgres, err := config.NewPostgres(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	d, err := delivery.New(usecase.New(repository.New(postgres)))
	if err != nil {
		log.Fatal().Err(err).Msg("can't create delivery")
	}

	mux := config.NewMux()
	mux.Get("/", d.Index)
	mux.Get("/list_games", d.ListGames)
	mux.Get("/registration_form/{id}", d.RegistrationForm)
	mux.Post("/register", d.Register)

	fs := http.FileServer(http.Dir("./front"))
	mux.Handle("/static/*", http.StripPrefix("/static/", fs))

	log.Info().Str("port", ":8080").Msg("server starts")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal().Err(err).Msg("ListenAndServe()")
	}
}
