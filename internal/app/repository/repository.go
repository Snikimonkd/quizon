package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"

	"quizon/internal/app/usecase"
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

// Transactional - выполняет fn в одной транзакции: коммит, если fn вернула nil, иначе откат
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
		return "", usecase.ErrNotFound
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
		table.Registrations.GameID.EQ(postgres.Int64(gameID)).
			AND(table.Registrations.DeletedAt.IS_NULL()),
	)

	query, args := stmt.Sql()
	var res *int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&res)
	if err != nil {
		return 0, fmt.Errorf("can't get registrations amount: %w", err)
	}

	return lo.FromPtr(res), nil
}

// GetGame - игра по id. usecase.ErrNotFound, если такой нет
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
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Games{}, usecase.ErrNotFound
	}
	if err != nil {
		return model.Games{}, fmt.Errorf("can't get game: %w", err)
	}

	return res, nil
}

// ListRegistrations - все регистрации на игру в порядке подачи, удалённые тоже
func (r repository) ListRegistrations(ctx context.Context, gameID int64) ([]model.Registrations, error) {
	stmt := table.Registrations.SELECT(
		table.Registrations.AllColumns,
	).WHERE(
		table.Registrations.GameID.EQ(postgres.Int64(gameID)),
	).ORDER_BY(
		table.Registrations.CreatedAt.ASC(),
		table.Registrations.ID.ASC(),
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
			&buf.ID,
			&buf.GameID,
			&buf.CreatedAt,
			&buf.TeamName,
			&buf.CaptainName,
			&buf.Phone,
			&buf.Telegram,
			&buf.PlayersAmount,
			&buf.GroupName,
			&buf.TeamID,
			&buf.DeletedAt,
		)

		if rErr != nil {
			return nil, fmt.Errorf("can't scan games: %w", rErr)
		}
		res = append(res, buf)
	}
	rErr := rows.Err()
	if rErr != nil {
		return nil, fmt.Errorf("error while scanning: %w", rErr)
	}

	return res, nil
}

// LockGame - блокирует игру до конца транзакции tx: кто заблокировал игру позже, ждёт её конца
func (r repository) LockGame(ctx context.Context, tx pgx.Tx, gameID int64) error {
	query := `select pg_advisory_xact_lock($1);`
	_, err := tx.Exec(ctx, query, gameID)
	if err != nil {
		return fmt.Errorf("can't lock game: %w", err)
	}

	return nil
}

// GetGameWithRegistrations - игра и число оставшихся в ней команд в транзакции tx.
// usecase.ErrNotFound, если игры нет
func (r repository) GetGameWithRegistrations(
	ctx context.Context,
	tx pgx.Tx,
	gameID int64,
) (usecase.GameWithRegistrations, error) {
	query := `
    SELECT g.*, (
        SELECT COUNT(1) FROM registrations
        WHERE game_id = g.id AND deleted_at IS NULL
    ) AS cnt
    FROM games g
    WHERE g.id = $1;
    `

	var res usecase.GameWithRegistrations
	err := tx.QueryRow(ctx, query, gameID).Scan(
		&res.ID,
		&res.CreatedAt,
		&res.StartTime,
		&res.Location,
		&res.Name,
		&res.MainAmount,
		&res.ReserveAmount,
		&res.RegistrationOpenTime,
		&res.RegistrationsAmount,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return usecase.GameWithRegistrations{}, usecase.ErrNotFound
	}
	if err != nil {
		return usecase.GameWithRegistrations{}, fmt.Errorf("can't get game with registrations: %w", err)
	}

	return res, nil
}

// CreateRegistration - сохраняет регистрацию в транзакции tx
func (r repository) CreateRegistration(ctx context.Context, tx pgx.Tx, in model.Registrations) error {
	// id выдаёт база, новая регистрация удалённой не бывает
	stmt := table.Registrations.INSERT(
		table.Registrations.AllColumns.Except(table.Registrations.ID, table.Registrations.DeletedAt),
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

// DeleteRegistration - помечает регистрацию удалённой в транзакции tx.
// usecase.ErrNotFound, если её нет в этой игре или она уже удалена
func (r repository) DeleteRegistration(
	ctx context.Context,
	tx pgx.Tx,
	gameID int64,
	registrationID int64,
	deletedAt time.Time,
) error {
	stmt := table.Registrations.UPDATE(
		table.Registrations.DeletedAt,
	).SET(
		postgres.TimestampzT(deletedAt),
	).WHERE(
		table.Registrations.ID.EQ(postgres.Int64(registrationID)).
			AND(table.Registrations.GameID.EQ(postgres.Int64(gameID))).
			AND(table.Registrations.DeletedAt.IS_NULL()),
	)

	query, args := stmt.Sql()
	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("can't delete registration: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return usecase.ErrNotFound
	}

	return nil
}

func (r repository) ListGames(ctx context.Context, page int64, perPage int64) ([]usecase.GameWithRegistrations, error) {
	query := `
    SELECT g.*, COALESCE(r.cnt, 0) AS cnt
    FROM games g LEFT JOIN (
        SELECT game_id, COUNT(1) AS cnt
        FROM registrations
        WHERE deleted_at IS NULL
        GROUP BY game_id
    ) r
    ON g.id = r.game_id
    ORDER BY g.start_time DESC
    LIMIT $1 OFFSET $2;
    `

	var res []usecase.GameWithRegistrations
	rows, err := r.db.Query(ctx, query, perPage, (page-1)*perPage)
	if err != nil {
		return nil, fmt.Errorf("can't select games: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var buf usecase.GameWithRegistrations
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

func (r repository) rollbackUnlessCommited(ctx context.Context, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	if errors.Is(err, pgx.ErrTxClosed) {
		return
	}
	if err != nil {
		slog.ErrorContext(ctx, "can't rollback tx", "err", err)
	}
}
