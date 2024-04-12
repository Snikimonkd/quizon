package delivery

import (
	"context"
	"net/http"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/rs/zerolog/log"
)

type ListGamesUsecase interface {
	ListGames(ctx context.Context, limit int64, offset int64) ([]model.Game, error)
}

type ListGamesResponse struct {
	Games []model.Game
}

func (d delivery) Index(w http.ResponseWriter, r *http.Request) {
	games, err := d.listGamesUsecase.ListGames(r.Context(), 100, 0)
	if err != nil {
		log.Error().Err(err).Msg("")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := ListGamesResponse{Games: games}

	w.WriteHeader(http.StatusOK)
	err = d.templ.ExecuteTemplate(w, "index", resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't execute template")
		return
	}
}

func (d delivery) ListGames(w http.ResponseWriter, r *http.Request) {
	games, err := d.listGamesUsecase.ListGames(r.Context(), 100, 0)
	if err != nil {
		log.Error().Err(err).Msg("")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := ListGamesResponse{Games: games}

	w.WriteHeader(http.StatusOK)
	err = d.templ.ExecuteTemplate(w, "games", resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't execute template")
		return
	}
}
