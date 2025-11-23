package main

import (
	"context"
	"errors"
	"log"
	nethttp "net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/derletzte256/avito-assignment-2025-autumn/internal/config"
	delivery "github.com/derletzte256/avito-assignment-2025-autumn/internal/delivery/http"
	_ "github.com/derletzte256/avito-assignment-2025-autumn/migrations"
	"github.com/derletzte256/avito-assignment-2025-autumn/pkg/database"
	"github.com/derletzte256/avito-assignment-2025-autumn/pkg/logger"

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

	l := logger.Get()

	if err = database.RunMigrations(ctx, cfg.Database); err != nil {
		l.Fatal("run migrations", zap.Error(err))
	}

	pool, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		l.Fatal("init postgres pool", zap.Error(err))
	}
	defer pool.Close()

	trManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	router := delivery.NewRouter(pool, trManager)

	srv := delivery.NewServer(cfg.HTTP, router, l)

	go func() {
		if err = srv.Start(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			l.Fatal("http server start", zap.Error(err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		l.Fatal("http server shutdown", zap.Error(err))
	}
}
