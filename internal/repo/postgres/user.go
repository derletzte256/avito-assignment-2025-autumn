package postgres

import (
	"avito-assignment-2025-autumn/internal/entity"
	"context"
	"errors"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	checkUserExistsQuery = `
		SELECT 1 FROM "user" WHERE id = $1
		`
	createUserQuery = `
		INSERT INTO "user" (id, username, is_active, team_name) 
		VALUES ($1, $2, $3, $4)
		`
	getUsersByTeamNameQuery = `
		SELECT id, username, is_active, team_name 
		FROM "user" 
		WHERE team_name = $1
		`
	findExistingByIDsQuery = `
		SELECT id 
		FROM "user" 
		WHERE id = ANY($1)
		`
	updateUsersQuery = `
		UPDATE "user" 
		SET username = $2, is_active = $3, team_name = $4 
		WHERE id = $1
		`
	setIsActiveQuery = `
		UPDATE "user" 
		SET is_active = $2 
		WHERE id = $1
		`
	getUserByIDQuery = `
		SELECT id, username, is_active, team_name 
		FROM "user" 
		WHERE id = $1`
	getReviewersForPRQuery = `
		SELECT id, username, is_active, team_name 
		FROM "user" 
		WHERE team_name = $1 AND id <> $2 AND is_active = true 
		ORDER BY RANDOM() LIMIT 2`
)

type UserRepo struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
	logger *zap.Logger
}

func NewUserRepo(db *pgxpool.Pool, getter *trmpgx.CtxGetter, logger *zap.Logger) *UserRepo {
	return &UserRepo{db: db, getter: getter, logger: logger}
}

func (r *UserRepo) CheckUserExists(ctx context.Context, id string) (bool, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	var exists int
	err := conn.QueryRow(ctx, checkUserExistsQuery, id).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *UserRepo) CreateBatch(ctx context.Context, users []*entity.User) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	batch := pgx.Batch{}
	for _, user := range users {
		batch.Queue(createUserQuery, user.ID, user.Username, user.IsActive, user.TeamName)
	}

	br := conn.SendBatch(ctx, &batch)
	defer func(br pgx.BatchResults) {
		err := br.Close()
		if err != nil {
			fmt.Printf("error closing batch results: %v\n", err)
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

func (r *UserRepo) GetByTeamName(ctx context.Context, teamName string) ([]*entity.User, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getUsersByTeamNameQuery, teamName)
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
		err := rows.Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamName)
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

func (r *UserRepo) FindExistingByIDs(ctx context.Context, ids []string) (map[string]struct{}, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, findExistingByIDsQuery, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existingIDs := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		existingIDs[id] = struct{}{}
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return existingIDs, nil
}

func (r *UserRepo) UpdateUsers(ctx context.Context, users []*entity.User) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	batch := pgx.Batch{}
	for _, user := range users {
		batch.Queue(updateUsersQuery, user.ID, user.Username, user.IsActive, user.TeamName)
	}

	br := conn.SendBatch(ctx, &batch)
	defer func(br pgx.BatchResults) {
		err := br.Close()
		if err != nil {
			fmt.Printf("error closing batch results: %v\n", err)
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

func (r *UserRepo) SetIsActive(ctx context.Context, id string, isActive bool) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	_, err := conn.Exec(ctx, setIsActiveQuery, id, isActive)
	return err
}

func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	user := &entity.User{}
	err := conn.QueryRow(ctx, getUserByIDQuery, id).Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) GetReviewersForPullRequest(ctx context.Context, teamName, excludeUserID string) ([]*entity.User, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getReviewersForPRQuery, teamName, excludeUserID)
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
		err := rows.Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamName)
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
