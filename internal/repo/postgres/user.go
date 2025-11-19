package postgres

import (
	"avito-assignment-2025-autumn/internal/entity"
	"context"
	"errors"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createUserQuery       = `INSERT INTO "user" (id, username, is_active, team_id) VALUES ($1, $2, $3, $4)`
	getUsersByTeamIDQuery = `SELECT id, username, is_active, team_id FROM "user" WHERE team_id = $1`
)

type UserRepo struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewUserRepo(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *UserRepo {
	return &UserRepo{db: db, getter: getter}
}

func (r *UserRepo) CreateBatch(ctx context.Context, users []*entity.User) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	batch := pgx.Batch{}
	for _, user := range users {
		batch.Queue(createUserQuery, user.ID, user.Username, user.IsActive, user.TeamID)
	}

	br := conn.SendBatch(ctx, &batch)
	defer func(br pgx.BatchResults) {
		err := br.Close()
		if err != nil {
			panic(err)
		}
	}(br)

	for range users {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *UserRepo) GetByTeamID(ctx context.Context, teamID int64) ([]*entity.User, error) {
	conn := r.db

	rows, err := conn.Query(ctx, getUsersByTeamIDQuery, teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*entity.User{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)

	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamID)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return users, nil
}
