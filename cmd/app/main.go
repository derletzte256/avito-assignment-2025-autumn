package main

import (
	"avito-assignment-2025-autumn/internal/config"
	delivery "avito-assignment-2025-autumn/internal/delivery/http"
	_ "avito-assignment-2025-autumn/migrations"
	"avito-assignment-2025-autumn/pkg/database"
	"context"
	"errors"
	"log"
	nethttp "net/http"
	"os/signal"
	"syscall"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}

	if err := database.RunMigrations(ctx, cfg.Database); err != nil {
		logger.Fatal("run migrations", zap.Error(err))
	}

	pool, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatal("init postgres pool", zap.Error(err))
	}
	defer pool.Close()

	trManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	router := delivery.NewRouter(pool, trManager, logger)

	srv := delivery.NewServer(cfg.HTTP, router, logger)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			logger.Fatal("http server start", zap.Error(err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("http server shutdown", zap.Error(err))
	}
}
