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
		&res.ReserverAmount,
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
			&buf.TeamSize,
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

// func (r repository) UpsertAdmin(ctx context.Context, req model.Admins) error {
// 	stmt := table.Admins.INSERT(
// 		table.Admins.AllColumns,
// 	).MODEL(
// 		req,
// 	).ON_CONFLICT(
// 		table.Admins.ID,
// 	).DO_UPDATE(
// 		postgres.SET(
// 			table.Admins.DateUntil.SET(table.Admins.EXCLUDED.DateUntil),
// 		),
// 	)
//
// 	query, args := stmt.Sql()
// 	_, err := r.db.Exec(ctx, query, args...)
// 	if err != nil {
// 		return fmt.Errorf("can't upsert admin: %w", err)
// 	}
//
// 	return nil
// }
//
// func (r repository) CheckAuth(ctx context.Context, userID int64) (model.Admins, error) {
// 	stmt := table.Admins.SELECT(
// 		table.Admins.ID,
// 		table.Admins.DateUntil,
// 	).WHERE(
// 		table.Admins.ID.EQ(postgres.Int64(userID)),
// 	)
//
// 	query, args := stmt.Sql()
// 	var res model.Admins
// 	err := r.db.QueryRow(ctx, query, args...).Scan(&res.ID, &res.DateUntil)
// 	if errors.Is(err, pgx.ErrNoRows) {
// 		return model.Admins{}, usecase.ErrNotFound
// 	}
// 	if err != nil {
// 		return model.Admins{}, fmt.Errorf("can't select from admins: %w", err)
// 	}
//
// 	return res, nil
// }
//
// func (r repository) DeletDraft(ctx context.Context, userID int64) error {
// 	stmt := table.RegistrationsDraft.DELETE().WHERE(
// 		table.RegistrationsDraft.UserID.EQ(postgres.Int64(userID)),
// 	)
//
// 	query, args := stmt.Sql()
// 	_, err := r.db.Exec(ctx, query, args...)
// 	if err != nil {
// 		return fmt.Errorf("can't delete from registrations_draft: %w", err)
// 	}
//
// 	return nil
// }
//
// func (r repository) GetRegistrationDraft(
// 	ctx context.Context,
// 	tx pgx.Tx,
// 	userID int64,
// ) (model.RegistrationsDraft, error) {
// 	stmt := table.RegistrationsDraft.SELECT(
// 		table.RegistrationsDraft.AllColumns,
// 	).WHERE(
// 		table.RegistrationsDraft.UserID.EQ(postgres.Int64(userID)),
// 	)
//
// 	query, args := stmt.Sql()
// 	var res model.RegistrationsDraft
// 	err := tx.QueryRow(ctx, query, args...).Scan(
// 		&res.UserID,
// 		&res.TgContact,
// 		&res.TeamID,
// 		&res.TeamName,
// 		&res.CaptainName,
// 		&res.GroupName,
// 		&res.Phone,
// 		&res.Amount,
// 		&res.CreatedAt,
// 		&res.UpdatedAt,
// 	)
//
// 	if errors.Is(err, pgx.ErrNoRows) {
// 		return model.RegistrationsDraft{}, usecase.ErrNotFound
// 	}
// 	if err != nil {
// 		return model.RegistrationsDraft{}, fmt.Errorf(
// 			"can't select from registrations draft: %w",
// 			err,
// 		)
// 	}
//
// 	return res, nil
// }
//
// func (r repository) UpdateRegistrationDraft(
// 	ctx context.Context,
// 	tx pgx.Tx,
// 	in model.RegistrationsDraft,
// ) error {
// 	stmt := table.RegistrationsDraft.INSERT(
// 		table.RegistrationsDraft.AllColumns,
// 	).MODEL(
// 		in,
// 	).ON_CONFLICT(
// 		table.RegistrationsDraft.UserID,
// 	).DO_UPDATE(
// 		postgres.SET(
// 			table.RegistrationsDraft.TeamID.SET(table.RegistrationsDraft.EXCLUDED.TeamID),
// 			table.RegistrationsDraft.TeamName.SET(table.RegistrationsDraft.EXCLUDED.TeamName),
// 			table.RegistrationsDraft.CaptainName.SET(table.RegistrationsDraft.EXCLUDED.CaptainName),
// 			table.RegistrationsDraft.GroupName.SET(table.RegistrationsDraft.EXCLUDED.GroupName),
// 			table.RegistrationsDraft.Phone.SET(table.RegistrationsDraft.EXCLUDED.Phone),
// 			table.RegistrationsDraft.Amount.SET(table.RegistrationsDraft.EXCLUDED.Amount),
// 			table.RegistrationsDraft.UpdatedAt.SET(table.RegistrationsDraft.EXCLUDED.UpdatedAt),
// 		),
// 	)
//
// 	query, args := stmt.Sql()
// 	_, err := tx.Exec(ctx, query, args...)
// 	if err != nil {
// 		return fmt.Errorf("can't update registrations_draft: %w", err)
// 	}
//
// 	return nil
// }

func (r repository) GenerateTeamID(ctx context.Context, tx pgx.Tx) (int64, error) {
	query := `SELECT nextval('team_id_seq');`
	var res int64
	err := tx.QueryRow(ctx, query).Scan(&res)
	if err != nil {
		return 0, fmt.Errorf("can't generate team_id: %w", err)
	}

	return res, nil
}
