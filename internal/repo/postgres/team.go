package postgres

import (
	"avito-assignment-2025-autumn/internal/entity"
	"context"
	"database/sql"
	"errors"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	checkTeamNameExistsQuery = `
		SELECT 1 
		FROM team 
		WHERE name = $1
		`
	createTeamQuery = `
		INSERT INTO team (name) 
		VALUES ($1)
		`
	getTeamByNameQuery = `
		SELECT name 
		FROM team 
		WHERE name = $1
		`
)

type TeamRepo struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
	logger *zap.Logger
}

func NewTeamRepo(db *pgxpool.Pool, getter *trmpgx.CtxGetter, logger *zap.Logger) *TeamRepo {
	return &TeamRepo{db: db, getter: getter, logger: logger}
}

func (r *TeamRepo) CheckTeamNameExists(ctx context.Context, name string) (bool, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	var exists int
	err := conn.QueryRow(ctx, checkTeamNameExistsQuery, name).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *TeamRepo) Create(ctx context.Context, team *entity.Team) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	_, err := conn.Exec(ctx, createTeamQuery, team.Name)
	return err
}

func (r *TeamRepo) GetByName(ctx context.Context, name string) (*entity.Team, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	team := &entity.Team{}
	err := conn.QueryRow(ctx, getTeamByNameQuery, name).Scan(&team.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return team, nil
}
