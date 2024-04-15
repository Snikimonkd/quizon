package repository

import (
	"context"
	"fmt"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
)

// ListGames - вывести игры
func (r repository) ListGames(ctx context.Context, limit int64, offset int64) ([]model.Game, error) {
	query := ` 
    SELECT g.id,
           g.location,
           g.start_time,
           g.name,
           g.teams_amount,
           g.reserve,
           g.registration_start,
           g.comment,
           g.created_at,
           g.updated_at,
           count(r.team_id)
    FROM games g LEFT JOIN registrations r ON g.id = r.game_id
    GROUP BY g.id
    ORDER BY start_time DESC
    LIMIT $1
    OFFSET $2;
    `

	res := make([]model.Game, 0, limit)
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("can't select from games: %w", err)
	}
	for rows.Next() {
		var buf model.Game
		cerr := rows.Scan(
			&buf.ID,
			&buf.Location,
			&buf.StartTime,
			&buf.Name,
			&buf.TeamsAmount,
			&buf.Reserve,
			&buf.RegistrationStart,
			&buf.Comment,
			&buf.CreatedAt,
			&buf.UpdatedAt,
			&buf.RegisteredTeams,
		)
		if cerr != nil {
			return nil, fmt.Errorf("can't scan game: %w", cerr)
		}

		res = append(res, buf)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error while scanning: %w", err)
	}

	return res, nil
}
