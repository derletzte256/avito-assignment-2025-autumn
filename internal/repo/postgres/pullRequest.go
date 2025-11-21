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
	checkPullRequestIDExistsQuery = `
		SELECT 1 FROM pull_request WHERE id = $1
		`
	createPullRequestQuery = `
		INSERT INTO pull_request (id, name, author_id, status_id) 
		VALUES ($1, $2, $3, (SELECT id FROM pull_request_status WHERE name = $4))
		`
	assignReviewersQuery = `
		INSERT INTO reviewer (pull_request_id, user_id) VALUES ($1, $2)
		`
	getPullRequestQuery = `
		SELECT pr.id, pr.name, pr.author_id, prs.name, pr.merged_at FROM pull_request pr 
        JOIN pull_request_status prs ON pr.status_id = prs.id WHERE pr.id = $1
        `
	getReviewersByPullRequestIDQuery = `
		SELECT user_id FROM reviewer WHERE pull_request_id = $1
		`
	getPullRequestsByReviewerIDQuery = `
		SELECT pr.id, pr.name, pr.author_id, prs.name, pr.merged_at FROM pull_request pr
		JOIN pull_request_status prs ON pr.status_id = prs.id
		JOIN reviewer r ON pr.id = r.pull_request_id
		WHERE r.user_id = $1
		ORDER BY pr.created_at DESC
		`
	mergePullRequestByIDQuery = `
		UPDATE pull_request 
		SET status_id = (SELECT id FROM pull_request_status WHERE name = 'MERGED'), merged_at = NOW() 
		WHERE id = $1
		`
)

type PullRequestRepo struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
	logger *zap.Logger
}

func NewPullRequestRepo(db *pgxpool.Pool, getter *trmpgx.CtxGetter, logger *zap.Logger) *PullRequestRepo {
	return &PullRequestRepo{db: db, getter: getter, logger: logger}
}

func (r *PullRequestRepo) CheckPullRequestIDExists(ctx context.Context, id string) (bool, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	var exists int
	err := conn.QueryRow(ctx, checkPullRequestIDExistsQuery, id).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *PullRequestRepo) Create(ctx context.Context, pr *entity.PullRequest) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	_, err := conn.Exec(ctx, createPullRequestQuery, pr.ID, pr.Name, pr.AuthorID, pr.Status)
	return err
}

func (r *PullRequestRepo) AssignReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	batch := pgx.Batch{}
	for _, reviewerID := range reviewerIDs {
		batch.Queue(assignReviewersQuery, prID, reviewerID)
	}

	br := conn.SendBatch(ctx, &batch)
	defer func(br pgx.BatchResults) {
		err := br.Close()
		if err != nil {
			fmt.Printf("error closing batch results: %v\n", err)
		}
	}(br)

	for range reviewerIDs {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PullRequestRepo) GetPullRequestByID(ctx context.Context, id string) (*entity.PullRequest, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	pr := &entity.PullRequest{}
	err := conn.QueryRow(ctx, getPullRequestQuery, id).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.MergedAt)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (r *PullRequestRepo) GetReviewersByPullRequestID(ctx context.Context, id string) ([]string, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getReviewersByPullRequestIDQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviewerIDs := make([]string, 0)

	for rows.Next() {
		var reviewerID string
		err := rows.Scan(&reviewerID)
		if err != nil {
			return nil, err
		}
		reviewerIDs = append(reviewerIDs, reviewerID)
	}

	return reviewerIDs, nil
}

func (r *PullRequestRepo) GetPullRequestsByReviewerID(ctx context.Context, userID string) ([]*entity.PullRequest, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getPullRequestsByReviewerIDQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pullRequests := make([]*entity.PullRequest, 0)

	for rows.Next() {
		pr := &entity.PullRequest{}
		err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.MergedAt)
		if err != nil {
			return nil, err
		}
		pullRequests = append(pullRequests, pr)
	}

	return pullRequests, nil
}

func (r *PullRequestRepo) MergePullRequestByID(ctx context.Context, id string) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	_, err := conn.Exec(ctx, mergePullRequestByIDQuery, id)
	return err
}
