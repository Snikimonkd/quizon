package delivery

import (
	"context"
	"net/http"
	"time"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/rs/zerolog/log"
)

type ListGamesUsecase interface {
	ListGames(ctx context.Context, limit int64, offset int64) ([]model.Game, error)
}

type GameEntry struct {
	model.Game
	ButtonText string
	IsActive   bool
}

type ListGamesResponse struct {
	Games []GameEntry
}

func (d delivery) Index(w http.ResponseWriter, r *http.Request) {
	games, err := d.listGamesUsecase.ListGames(r.Context(), 100, 0)
	if err != nil {
		log.Error().Err(err).Msg("")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	entries := make([]GameEntry, 0, len(games))
	now := time.Now()
	for _, game := range games {
		entry := GameEntry{
			Game: game,
		}
		if game.RegistrationStart.After(now) {
			entry.ButtonText = "Регистарция скоро откроется"
			entry.IsActive = false
		} else if now.After(game.StartTime) {
			entry.ButtonText = "Регистарция закрыта"
			entry.IsActive = false
		} else if entry.RegisteredTeams >= entry.TeamsAmount+entry.Reserve {
			entry.ButtonText = "Регистарция закрыта"
			entry.IsActive = false
		} else if entry.RegisteredTeams >= entry.TeamsAmount {
			entry.ButtonText = "Зарегестрироваться в резерв"
			entry.IsActive = true
		} else {
			entry.ButtonText = "Зарегестрироваться"
			entry.IsActive = true
		}

		entries = append(entries, entry)
	}

	resp := ListGamesResponse{Games: entries}

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

	entries := make([]GameEntry, 0, len(games))
	now := time.Now()
	for _, game := range games {
		entry := GameEntry{
			Game: game,
		}
		if game.RegistrationStart.After(now) {
			entry.ButtonText = "Регистарция скоро откроется"
			entry.IsActive = false
		} else if now.After(game.StartTime) {
			entry.ButtonText = "Регистарция закрыта"
			entry.IsActive = false
		} else if entry.RegisteredTeams >= entry.TeamsAmount+entry.Reserve {
			entry.ButtonText = "Регистарция закрыта"
			entry.IsActive = false
		} else if entry.RegisteredTeams >= entry.TeamsAmount {
			entry.ButtonText = "Зарегестрироваться в резерв"
			entry.IsActive = true
		} else {
			entry.ButtonText = "Зарегестрироваться"
			entry.IsActive = true
		}

		entries = append(entries, entry)
	}

	resp := ListGamesResponse{Games: entries}

	w.WriteHeader(http.StatusOK)
	err = d.templ.ExecuteTemplate(w, "games", resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't execute template")
		return
	}
}
