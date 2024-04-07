package repository

import (
	"context"
	"fmt"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
)

// CreateGame - создать игру
func (r repository) CreateGame(ctx context.Context, game model.Game) error {
	query := `
    INSERT INTO games (
        location,
        start_time,
        name,
        teams_amount,
        reserve,
        registration_start,
        comment,
        created_at,
        updated_at
    ) VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9
    );
    `

	_, err := r.db.Exec(
		ctx,
		query,
		game.Location,
		game.StartTime,
		game.TeamsAmount,
		game.Reserve,
		game.RegistrationStart,
		game.Comment,
		game.CreatedAt,
		game.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("can't insert into games: %w", err)
	}

	return nil
}
