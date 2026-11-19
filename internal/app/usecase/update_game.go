package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"

	"quizon/internal/generated/postgres/public/model"
)

type UpdateGameRepository interface {
	UpdateGame(ctx context.Context, tx pgx.Tx, gameID int64, in model.Games) error
}

// UpdateGame - обновляет настройки игры под той же блокировкой, что и регистрация команд.
// ValidationError, если параметры нарушают правила, ErrNotFound, если игры нет.
func (u usecase) UpdateGame(ctx context.Context, gameID int64, in model.Games) error {
	err := validateGame(in)
	if err != nil {
		return err
	}

	return u.repository.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
		txErr := u.repository.LockGame(ctx, tx, gameID)
		if txErr != nil {
			return txErr
		}

		return u.repository.UpdateGame(ctx, tx, gameID, in)
	})
}
