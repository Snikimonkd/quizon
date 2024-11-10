package repository

import (
	"context"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"

	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/generated/postgres/public/table"
)

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) repository {
	return repository{
		db: db,
	}
}

func (r repository) RegistrationsAmount(ctx context.Context, gameID int64) (int64, error) {
	stmt := table.Registrations.SELECT(
		postgres.COUNT(postgres.STAR),
	).WHERE(
		table.Registrations.GameID.EQ(postgres.Int64(gameID)),
	)

	query, args := stmt.Sql()
	var res *int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&res)
	if err != nil {
		return 0, fmt.Errorf("can't get registrations amount: %w", err)
	}

	return lo.FromPtr(res), nil
}

func (r repository) GetGame(ctx context.Context, gameID int64) (model.Games, error) {
	stmt := table.Games.SELECT(
		table.Games.AllColumns,
	).WHERE(
		table.Games.ID.EQ(postgres.Int64(gameID)),
	)

	query, args := stmt.Sql()
	var res model.Games
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&res.ID,
		&res.CreatedAt,
		&res.StartTime,
		&res.Location,
		&res.Name,
		&res.MainAmount,
		&res.ReserveAmount,
		&res.RegistartionOpenTime,
	)
	if err != nil {
		return model.Games{}, fmt.Errorf("can't get registration restrictions: %w", err)
	}

	return res, nil
}

func (r repository) Registrations(ctx context.Context, gameID int64) ([]model.Registrations, error) {
	stmt := table.Registrations.SELECT(
		table.Registrations.AllColumns,
	).WHERE(
		table.Registrations.GameID.EQ(postgres.Int64(gameID)),
	).ORDER_BY(
		table.Registrations.CreatedAt.ASC(),
	)

	query, args := stmt.Sql()
	var res []model.Registrations
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("can't select from registrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var buf model.Registrations
		rErr := rows.Scan(
			&buf.GameID,
			&buf.CreatedAt,
			&buf.TeamName,
			&buf.CaptainName,
			&buf.Phone,
			&buf.Telegram,
			&buf.PlayersAmount,
			&buf.GroupName,
			&buf.TeamID,
		)

		if rErr != nil {
			return nil, fmt.Errorf("can't scan games: %w", rErr)
		}
		res = append(res, buf)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("error while scanning: %w", err)
	}

	return res, nil
}

func (r repository) Register(ctx context.Context, in model.Registrations) error {
	stmt := table.Registrations.INSERT(
		table.Registrations.AllColumns,
	).MODEL(
		in,
	)

	query, args := stmt.Sql()
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("can't insert into registrations: %w", err)
	}

	return nil
}

func (r repository) CheckTeamsAmount(ctx context.Context, tx pgx.Tx) (int64, error) {
	stmt := table.Registrations.SELECT(
		postgres.COUNT(postgres.STAR),
	)

	query, args := stmt.Sql()
	var res *int64
	err := tx.QueryRow(ctx, query, args...).Scan(&res)
	if err != nil {
		return 0, fmt.Errorf("can't select max teams amount: %w", err)
	}

	return lo.FromPtr(res), nil
}

func (r repository) ListGames(ctx context.Context) ([]model.Games, error) {
	stmt := table.Games.SELECT(
		table.Games.AllColumns,
	).ORDER_BY(
		table.Games.StartTime,
	)

	var res []model.Games
	query, args := stmt.Sql()
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("can't select games: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var buf model.Games
		cerr := rows.Scan(
			&buf.ID,
			&buf.CreatedAt,
			&buf.StartTime,
			&buf.Location,
			&buf.Name,
			&buf.MainAmount,
			&buf.ReserveAmount,
			&buf.RegistartionOpenTime,
		)
		if cerr != nil {
			return nil, fmt.Errorf("can't scan game: %w", cerr)
		}
		res = append(res, buf)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("error while scanning games: %w", rows.Err())
	}

	return res, nil
}

func (r repository) CreateGame(ctx context.Context, game model.Games) error {
	stmt := table.Games.INSERT(
		table.Games.AllColumns.Except(table.Games.ID),
	).MODEL(
		game,
	)
	query, args := stmt.Sql()
	_, err := r.db.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("can't insert into games: %w", err)
	}

	return nil
}
