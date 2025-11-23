package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/jackc/pgx/v5"
)

const (
	defaultDBConnectTimeout = 30 * time.Second
)

func newExpect(t *testing.T) *httpexpect.Expect {
	t.Helper()

	baseURL := os.Getenv("SERVER_ADDRESS")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return httpexpect.Default(t, baseURL)
}

func setupPostgres() (func(), error) {
	dsn, err := buildPostgresDSNFromEnv()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultDBConnectTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	cleanup := func() {
		_ = conn.Close(context.Background())
	}

	if err = truncateTables(ctx, conn); err != nil {
		cleanup()
		return nil, err
	}

	return cleanup, nil
}

func buildPostgresDSNFromEnv() (string, error) {
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	name := os.Getenv("DATABASE_NAME")

	if host == "" || port == "" || user == "" || password == "" || name == "" {
		return "", fmt.Errorf("database environment not set)")
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, name), nil
}

func truncateTables(ctx context.Context, conn *pgx.Conn) error {
	const q = `
		TRUNCATE TABLE
			reviewer,
			pull_request,
			"user",
			team
		RESTART IDENTITY CASCADE;
	`

	if _, err := conn.Exec(ctx, q); err != nil {
		return fmt.Errorf("truncate tables: %w", err)
	}
	return nil
}
