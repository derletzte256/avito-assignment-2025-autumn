package team

import (
	"context"
	"errors"

	"github.com/derletzte256/avito-assignment-2025-autumn/internal/entity"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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

type Repo struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewTeamRepo(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repo {
	return &Repo{db: db, getter: getter}
}

func (r *Repo) CheckTeamNameExists(ctx context.Context, name string) (bool, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	var exists int
	err := conn.QueryRow(ctx, checkTeamNameExistsQuery, name).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repo) Create(ctx context.Context, team *entity.Team) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	_, err := conn.Exec(ctx, createTeamQuery, team.Name)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return entity.ErrAlreadyExists
		}
	}
	return err
}

func (r *Repo) GetByName(ctx context.Context, name string) (*entity.Team, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	team := &entity.Team{}
	err := conn.QueryRow(ctx, getTeamByNameQuery, name).Scan(&team.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return team, nil
}
