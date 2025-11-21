package http

import (
	"avito-assignment-2025-autumn/internal/delivery/http/handlers"
	"avito-assignment-2025-autumn/internal/repo/postgres"
	"avito-assignment-2025-autumn/internal/usecase"
	"net/http"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewRouter(pool *pgxpool.Pool, trManager *manager.Manager, logger *zap.Logger) http.Handler {
	r := mux.NewRouter()

	teamRepo := postgres.NewTeamRepo(pool, trmpgx.DefaultCtxGetter, logger)
	userRepo := postgres.NewUserRepo(pool, trmpgx.DefaultCtxGetter, logger)
	pullRequestRepo := postgres.NewPullRequestRepo(pool, trmpgx.DefaultCtxGetter, logger)

	teamUC := usecase.NewTeamUseCase(teamRepo, userRepo, trManager, logger)
	teamDelivery := handlers.NewTeamDelivery(teamUC, logger)

	userUC := usecase.NewUserUseCase(userRepo, pullRequestRepo, trManager, logger)
	userDelivery := handlers.NewUserDelivery(userUC, logger)

	pullRequestUC := usecase.NewPullRequestUseCase(pullRequestRepo, userRepo, teamRepo, trManager, logger)
	pullRequestDelivery := handlers.NewPullRequestDelivery(pullRequestUC, logger)

	teamSubrouter := r.PathPrefix("/team").Subrouter()
	userSubrouter := r.PathPrefix("/users").Subrouter()
	prSubrouter := r.PathPrefix("/pullRequest").Subrouter()

	teamDelivery.RegisterRoutes(teamSubrouter)
	userDelivery.RegisterRoutes(userSubrouter)
	pullRequestDelivery.RegisterRoutes(prSubrouter)

	return r
}
