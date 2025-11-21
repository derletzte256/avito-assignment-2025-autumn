package database

import (
	"avito-assignment-2025-autumn/internal/config"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func ConnString(cfg config.DatabaseConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}

func RunMigrations(ctx context.Context, cfg config.DatabaseConfig) error {
	connString := ConnString(cfg)

	sqlDB, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("open sql db: %w", err)
	}
	defer func(sqlDB *sql.DB) {
		err := sqlDB.Close()
		if err != nil {
			fmt.Printf("close sql db: %v", err)
		}
	}(sqlDB)

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping sql db: %w", err)
	}

	if err := goose.UpContext(ctx, sqlDB, "."); err != nil {
		return fmt.Errorf("apply goose migrations: %w", err)
	}

	return nil
}

func NewPostgresPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	connString := ConnString(cfg)

	poolCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}

	p, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := p.Ping(ctx); err != nil {
		p.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return p, nil
}
