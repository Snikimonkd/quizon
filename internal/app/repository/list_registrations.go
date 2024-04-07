package repository

import (
	"context"
	"fmt"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
)

func (r repository) ListRegistrations(ctx context.Context, gameID int64) ([]model.Registration, error) {
	query := `
    SELECT id,
           game_id,
           team_id,
           captain_name,
           captain_group,
           captain_telegram,
           team_name,
           team_size,
           created_at
    FROM registrations
    WHERE game_id = $1 AND deleted = false
    ORDER BY created_at ASC;
    `

	var res []model.Registration
	rows, err := r.db.Query(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("can't select from registrations: %w", err)
	}
	for rows.Next() {
		var buf model.Registration
		cerr := rows.Scan(
			&buf.ID,
			&buf.GameID,
			&buf.TeamID,
			&buf.CaptainName,
			&buf.CaptainGroup,
			&buf.CaptainTelegram,
			&buf.TeamName,
			&buf.TeamSize,
			&buf.CreatedAt,
		)
		if cerr != nil {
			return nil, fmt.Errorf("can't scan registration: %w", err)
		}

		res = append(res, buf)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error while scanning from registrations: %w", err)
	}

	return res, nil
}
