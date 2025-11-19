package main

import (
	"avito-assignment-2025-autumn/internal/config"
	delivery "avito-assignment-2025-autumn/internal/delivery/http"
	"avito-assignment-2025-autumn/internal/delivery/http/handlers"
	"avito-assignment-2025-autumn/internal/repo/postgres"
	"avito-assignment-2025-autumn/internal/usecase"
	_ "avito-assignment-2025-autumn/migrations"
	"avito-assignment-2025-autumn/pkg/database"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	nethttp "net/http"
	"os/signal"
	"syscall"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := runMigrations(ctx, cfg.Database); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	pool, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	trManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	teamRepo := postgres.NewTeamRepo(pool, trmpgx.DefaultCtxGetter)
	userRepo := postgres.NewUserRepo(pool, trmpgx.DefaultCtxGetter)
	teamUC := usecase.NewTeamUseCase(teamRepo, userRepo, trManager)
	teamDelivery := handlers.NewTeamDelivery(teamUC)

	srv := delivery.NewServer(cfg.HTTP, teamDelivery, nil, nil)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}

func runMigrations(ctx context.Context, cfg config.DatabaseConfig) error {
	connString := database.ConnString(cfg)

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
