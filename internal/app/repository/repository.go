package repository

import (
	"context"
	"errors"
	"fmt"

	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/generated/postgres/public/table"
	"quizon/internal/pkg/logger"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"
)

var ErrNotFound error = errors.New("not found")

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) repository {
	return repository{
		db: db,
	}
}

func (r repository) rollbackUnlessCommited(ctx context.Context, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	if errors.Is(err, pgx.ErrTxClosed) {
		return
	}
	if err != nil {
		logger.Errorf("can't rollback tx in defer: %v", err)
	}
}

func (r repository) Transactional(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can't begin tx: %w", err)
	}
	defer r.rollbackUnlessCommited(ctx, tx)

	err = fn(ctx, tx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("can't commit tx: %w", err)
	}

	return nil
}

func (r repository) GetPassword(ctx context.Context, login string) (string, error) {
	stmt := table.Admins.SELECT(
		table.Admins.Password,
	).WHERE(
		table.Admins.Login.EQ(postgres.String(login)),
	)

	var password string
	query, args := stmt.Sql()
	err := r.db.QueryRow(ctx, query, args...).Scan(&password)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("can't get password: %w", err)
	}

	return password, nil
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
		&res.RegistrationOpenTime,
	)
	if err != nil {
		return model.Games{}, fmt.Errorf("can't get registration restrictions: %w", err)
	}

	return res, nil
}

func (r repository) ListRegistrations(ctx context.Context, gameID int64) ([]model.Registrations, error) {
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

func (r repository) LockGame(ctx context.Context, tx pgx.Tx, gameID int64) error {
	query := `select pg_advisory_xact_lock($1);`
	_, err := tx.Exec(ctx, query, gameID)
	if err != nil {
		return fmt.Errorf("can't lock game: %w", err)
	}

	return nil
}

func (r repository) GetRegistrationsAmount(ctx context.Context, tx pgx.Tx, gameID int64) (int64, int64, int64, error) {
	query := `
    SELECT g.main_amount, g.reserve_amount, r.cnt
    FROM games g JOIN (
        SELECT $1::bigint AS game_id, COALESCE(COUNT(1), 0) AS cnt
        FROM registrations
        WHERE game_id = $1::bigint
    ) r
    ON g.id = r.game_id;
    `
	var amount, reserve, cnt int64
	err := tx.QueryRow(ctx, query, gameID).Scan(&amount, &reserve, &cnt)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("can't check game registrations: %w", err)
	}

	return amount, reserve, cnt, nil
}

func (r repository) CreateRegistration(ctx context.Context, tx pgx.Tx, in model.Registrations) error {
	stmt := table.Registrations.INSERT(
		table.Registrations.AllColumns,
	).MODEL(
		in,
	)

	query, args := stmt.Sql()
	_, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("can't insert into registrations: %w", err)
	}

	return nil
}

type GameWithRegistrations struct {
	model.Games
	RegistrationsAmount int64
}

func (r repository) ListGames(ctx context.Context, page int64, perPage int64) ([]GameWithRegistrations, error) {
	query := `
    SELECT g.*, COALESCE(r.cnt, 0) AS cnt
    FROM games g LEFT JOIN (
        SELECT game_id, COUNT(1) AS cnt
        FROM registrations
        GROUP BY game_id
    ) r
    ON g.id = r.game_id
    ORDER BY g.start_time DESC
    LIMIT $1 OFFSET $2;
    `

	var res []GameWithRegistrations
	rows, err := r.db.Query(ctx, query, perPage, (page-1)*perPage)
	if err != nil {
		return nil, fmt.Errorf("can't select games: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var buf GameWithRegistrations
		cerr := rows.Scan(
			&buf.ID,
			&buf.CreatedAt,
			&buf.StartTime,
			&buf.Location,
			&buf.Name,
			&buf.MainAmount,
			&buf.ReserveAmount,
			&buf.RegistrationOpenTime,
			&buf.RegistrationsAmount,
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

func (r repository) CreateGame(ctx context.Context, game model.Games) (int64, error) {
	stmt := table.Games.INSERT(
		table.Games.AllColumns.Except(table.Games.ID),
	).MODEL(
		game,
	).RETURNING(
		table.Games.ID,
	)

	var res int64
	query, args := stmt.Sql()
	err := r.db.QueryRow(ctx, query, args...).Scan(&res)
	if err != nil {
		return 0, fmt.Errorf("can't insert into games: %w", err)
	}

	return res, nil
}
