package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/rs/zerolog/log"
)

type ListGamesRepository interface {
	ListGames(ctx context.Context, limit int64, offset int64) ([]model.Game, error)
}

func weekdayToRus(d time.Weekday) string {
	switch d {
	case 0:
		return "ВОСКРЕСЕНЬЕ"
	case 1:
		return "ПОНЕДЕЛЬНИК"
	case 2:
		return "ВТОРНИК"
	case 3:
		return "СРЕДА"
	case 4:
		return "ЧЕТВЕРГ"
	case 5:
		return "ПЯТНИЦА"
	case 6:
		return "СУББОТА"
	}
	return ""
}

func monthToRus(m time.Month) string {
	switch m {
	case time.January:
		return "ЯНВАРЯ"
	case time.February:
		return "ФЕВРАЛЯ"
	case time.March:
		return "МАРТА"
	case time.April:
		return "АПРЕЛЯ"
	case time.May:
		return "МАЯ"
	case time.June:
		return "ИЮНЯ"
	case time.July:
		return "ИЮЛЯ"
	case time.August:
		return "АВГУСТА"
	case time.September:
		return "СЕНТЯБРЯ"
	case time.October:
		return "ОКТЯБРЯ"
	case time.November:
		return "НОЯБРЯ"
	case time.December:
		return "ДЕКАБРЯ"
	}
	return ""
}

func format(t time.Time) string {
	month := monthToRus(t.Month())
	day := t.Day()
	return fmt.Sprintf("%d %s", day, month)
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

		entry.Date = format(entry.StartTime)
		entry.Weekday = weekdayToRus(entry.StartTime.Weekday())

		entries = append(entries, entry)
	}

	return model.ListGamesResponse{Games: entries}, nil
}
