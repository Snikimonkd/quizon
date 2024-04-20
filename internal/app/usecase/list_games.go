package usecase

import (
	"context"
	"time"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/rs/zerolog/log"
)

type ListGamesRepository interface {
	ListGames(ctx context.Context, limit int64, offset int64) ([]model.Game, error)
}

func (u usecase) ListGames(ctx context.Context, limit int64, offset int64) (model.ListGamesResponse, error) {
	games, err := u.listGamesRepository.ListGames(ctx, limit, offset)
	if err != nil {
		return model.ListGamesResponse{}, err
	}

	entries := make([]model.GameEntry, 0, len(games))
	now := time.Now()
	for _, game := range games {
		log.Info().Msgf("registered teams: %d", game.RegisteredTeams)
		entry := model.GameEntry{
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

	return model.ListGamesResponse{Games: entries}, nil
}
