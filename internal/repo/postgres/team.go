package postgres

import (
	"avito-assignment-2025-autumn/internal/entity"
	"context"
	"database/sql"
	"errors"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createTeamQuery    = `INSERT INTO team (name) VALUES ($1) RETURNING id`
	getTeamByNameQuery = `SELECT id, name FROM team WHERE name = $1`
)

type TeamRepo struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewTeamRepo(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *TeamRepo {
	return &TeamRepo{db: db, getter: getter}
}

func (r *TeamRepo) Create(ctx context.Context, team *entity.Team) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	var existingName string
	err := conn.QueryRow(ctx, getTeamByNameQuery, team.Name).Scan(&existingName)
	if err == nil {
		return entity.ErrAlreadyExists
	}

	return conn.QueryRow(ctx, createTeamQuery, team.Name).Scan(&team.ID)
}

func (r *TeamRepo) GetByName(ctx context.Context, name string) (*entity.Team, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	team := &entity.Team{}
	err := conn.QueryRow(ctx, getTeamByNameQuery, name).Scan(&team.ID, &team.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return team, nil
}
