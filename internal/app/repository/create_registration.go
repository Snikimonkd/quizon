package repository

import (
	"context"
	"fmt"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
)

func (r repository) Register(ctx context.Context, registration model.Registration) error {
	query := `
    INSERT INTO registrations (
        game_id,
        team_id,
        captain_name,
        captain_group,
        captain_telegram,
        team_name,
        team_size,
        created_at
    ) VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8
    );
    `

	_, err := r.db.Exec(
		ctx,
		query,
		registration.GameID,
		registration.TeamID,
		registration.CaptainName,
		registration.CaptainGroup,
		registration.CaptainTelegram,
		registration.TeamName,
		registration.TeamSize,
		registration.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("can't insert into registrations: %w", err)
	}

	return nil
}
