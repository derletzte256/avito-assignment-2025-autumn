package pullRequest

import (
	"context"
	"errors"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/derletzte256/avito-assignment-2025-autumn/internal/entity"
	"github.com/derletzte256/avito-assignment-2025-autumn/internal/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	checkPullRequestIDExistsQuery = `
		SELECT 1 
		FROM pull_request 
		WHERE id = $1
		`
	createPullRequestQuery = `
		INSERT INTO pull_request (id, name, author_id, status_id) 
		VALUES ($1, $2, $3, (SELECT id FROM pull_request_status WHERE name = $4))
		`
	assignReviewersQuery = `
		INSERT INTO reviewer (pull_request_id, user_id) 
		VALUES ($1, $2)
		`
	getPullRequestQuery = `
		SELECT pr.id, pr.name, pr.author_id, prs.name, pr.merged_at 
		FROM pull_request pr 
        JOIN pull_request_status prs ON pr.status_id = prs.id 
		WHERE pr.id = $1
        `
	getReviewersByPullRequestIDQuery = `
		SELECT user_id 
		FROM reviewer 
		WHERE pull_request_id = $1
		`
	getPullRequestsByReviewerIDQuery = `
		SELECT pr.id, pr.name, pr.author_id, prs.name, pr.merged_at 
		FROM pull_request pr
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
	removeReviewerQuery = `
		DELETE 
		FROM reviewer 
		WHERE pull_request_id = $1 AND user_id = $2
		`

	getReviewStatisticsQuery = `
		SELECT
			r.user_id,
			SUM(CASE WHEN prs.name = 'OPEN' THEN 1 ELSE 0 END) AS open_count,
			SUM(CASE WHEN prs.name = 'MERGED' THEN 1 ELSE 0 END) AS merged_count
		FROM reviewer r
		JOIN pull_request pr ON pr.id = r.pull_request_id
		JOIN pull_request_status prs ON prs.id = pr.status_id
		GROUP BY r.user_id
		`
	getAuthorStatisticsQuery = `
		SELECT pr.author_id, COUNT(*) AS authored_count
		FROM pull_request pr
		GROUP BY pr.author_id
		`
	getReviewsByReviewerIDsQuery = `
		SELECT r.pull_request_id, r.user_id
		FROM reviewer r
		JOIN pull_request pr ON pr.id = r.pull_request_id
		JOIN pull_request_status s ON s.id = pr.status_id
		WHERE r.user_id = ANY($1) AND s.name = 'OPEN'
		`
	getReviewersByPullRequestIDsQuery = `
		SELECT pull_request_id, user_id
		FROM reviewer
		WHERE pull_request_id = ANY($1)
		`
)

type Repo struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewRepo(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repo {
	return &Repo{db: db, getter: getter}
}

func (r *Repo) CheckPullRequestIDExists(ctx context.Context, id string) (bool, error) {
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

func (r *Repo) Create(ctx context.Context, pr *entity.PullRequest) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	_, err := conn.Exec(ctx, createPullRequestQuery, pr.ID, pr.Name, pr.AuthorID, pr.Status)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return entity.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *Repo) AssignReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	l := logger.FromCtx(ctx)

	batch := pgx.Batch{}
	for _, reviewerID := range reviewerIDs {
		batch.Queue(assignReviewersQuery, prID, reviewerID)
	}

	br := conn.SendBatch(ctx, &batch)
	defer func(br pgx.BatchResults) {
		err := br.Close()
		if err != nil {
			l.Error("error closing batch results: %v\n", zap.Error(err))
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

func (r *Repo) GetPullRequestByID(ctx context.Context, id string) (*entity.PullRequest, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	pr := &entity.PullRequest{}
	err := conn.QueryRow(ctx, getPullRequestQuery, id).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return pr, nil
}

func (r *Repo) GetReviewersByPullRequestID(ctx context.Context, id string) ([]string, error) {
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

func (r *Repo) GetPullRequestsByReviewerID(ctx context.Context, userID string) ([]*entity.PullRequest, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getPullRequestsByReviewerIDQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pullRequests := make([]*entity.PullRequest, 0)

	for rows.Next() {
		pr := &entity.PullRequest{}
		err = rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.MergedAt)
		if err != nil {
			return nil, err
		}
		pullRequests = append(pullRequests, pr)
	}

	return pullRequests, nil
}

func (r *Repo) MergePullRequestByID(ctx context.Context, id string) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	result, err := conn.Exec(ctx, mergePullRequestByIDQuery, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *Repo) RemoveReviewer(ctx context.Context, prID, userID string) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	result, err := conn.Exec(ctx, removeReviewerQuery, prID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *Repo) AddNewReviewer(ctx context.Context, prID, userID string) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	result, err := conn.Exec(ctx, assignReviewersQuery, prID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return entity.ErrNotFound
	}
	return err
}

func (r *Repo) GetOpenAndMergedReviewStatisticsForUsers(ctx context.Context) (map[string]int, map[string]int, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	openMap := make(map[string]int)
	mergedMap := make(map[string]int)

	rows, err := conn.Query(ctx, getReviewStatisticsQuery)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return openMap, mergedMap, nil
		}
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var openCount, mergedCount int

		if err = rows.Scan(&userID, &openCount, &mergedCount); err != nil {
			return nil, nil, err
		}

		openMap[userID] = openCount
		mergedMap[userID] = mergedCount
	}

	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}

	return openMap, mergedMap, nil
}

func (r *Repo) GetAuthorStatisticsForUsers(ctx context.Context) (map[string]int, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	authorMap := make(map[string]int)

	rows, err := conn.Query(ctx, getAuthorStatisticsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var authorID string
		var authoredCount int

		if err = rows.Scan(&authorID, &authoredCount); err != nil {
			return nil, err
		}
		authorMap[authorID] = authoredCount
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return authorMap, nil
}

func (r *Repo) GetReviewsByReviewerIDs(ctx context.Context, reviewerIDs []string) ([]*entity.ReviewRecord, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getReviewsByReviewerIDsQuery, reviewerIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*entity.ReviewRecord, 0)

	for rows.Next() {
		record := &entity.ReviewRecord{}
		if err = rows.Scan(&record.PullRequestID, &record.ReviewerID); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return records, nil
}

func (r *Repo) GetReviewersByPullRequestIDs(ctx context.Context, prIDs []string) (map[string][]string, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getReviewersByPullRequestIDsQuery, prIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]string)

	for rows.Next() {
		var prID string
		var reviewerID string
		if err = rows.Scan(&prID, &reviewerID); err != nil {
			return nil, err
		}
		result[prID] = append(result[prID], reviewerID)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (r *Repo) RemoveReviewersBatch(ctx context.Context, records []*entity.ReviewRecord) error {
	if len(records) == 0 {
		return nil
	}
	l := logger.FromCtx(ctx)

	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	batch := pgx.Batch{}
	for _, record := range records {
		batch.Queue(removeReviewerQuery, record.PullRequestID, record.ReviewerID)
	}

	br := conn.SendBatch(ctx, &batch)
	defer func(br pgx.BatchResults) {
		err := br.Close()
		if err != nil {
			l.Error("error closing batch results: %v\n", zap.Error(err))
		}
	}(br)

	for range records {
		result, err := br.Exec()
		if err != nil {
			return err
		}

		if result.RowsAffected() == 0 {
			return entity.ErrNotFound
		}
	}

	return nil
}

func (r *Repo) AddReviewersBatch(ctx context.Context, records []*entity.ReviewRecord) error {
	if len(records) == 0 {
		return nil
	}
	l := logger.FromCtx(ctx)

	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	batch := pgx.Batch{}
	for _, record := range records {
		batch.Queue(assignReviewersQuery, record.PullRequestID, record.ReviewerID)
	}

	br := conn.SendBatch(ctx, &batch)
	defer func(br pgx.BatchResults) {
		err := br.Close()
		if err != nil {
			l.Error("error closing batch results: %v\n", zap.Error(err))
		}
	}(br)

	for range records {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
