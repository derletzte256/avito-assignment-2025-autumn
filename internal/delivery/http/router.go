package http

import (
	prdelivery "avito-assignment-2025-autumn/internal/delivery/http/handlers/pullRequest"
	teamdelivery "avito-assignment-2025-autumn/internal/delivery/http/handlers/team"
	userdelivery "avito-assignment-2025-autumn/internal/delivery/http/handlers/user"
	prrepo "avito-assignment-2025-autumn/internal/repo/postgres/pullRequest"
	teamrepo "avito-assignment-2025-autumn/internal/repo/postgres/team"
	userrepo "avito-assignment-2025-autumn/internal/repo/postgres/user"
	prusecase "avito-assignment-2025-autumn/internal/usecase/pullRequest"
	teamusecase "avito-assignment-2025-autumn/internal/usecase/team"
	userusecase "avito-assignment-2025-autumn/internal/usecase/user"
	"net/http"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewRouter(pool *pgxpool.Pool, trManager *manager.Manager, logger *zap.Logger) http.Handler {
	r := mux.NewRouter()

	teamRepo := teamrepo.NewTeamRepo(pool, trmpgx.DefaultCtxGetter, logger)
	userRepo := userrepo.NewUserRepo(pool, trmpgx.DefaultCtxGetter, logger)
	pullRequestRepo := prrepo.NewRepo(pool, trmpgx.DefaultCtxGetter, logger)

	teamUC := teamusecase.NewUseCase(teamRepo, userRepo, trManager, logger)
	teamDelivery := teamdelivery.NewTeamDelivery(teamUC, logger)

	userUC := userusecase.NewUseCase(userRepo, pullRequestRepo, teamRepo, trManager, logger)
	userDelivery := userdelivery.NewUserDelivery(userUC, logger)

	pullRequestUC := prusecase.NewUseCase(pullRequestRepo, userRepo, teamRepo, trManager, logger)
	pullRequestDelivery := prdelivery.NewPullRequestDelivery(pullRequestUC, logger)

	teamSubrouter := r.PathPrefix("/team").Subrouter()
	userSubrouter := r.PathPrefix("/users").Subrouter()
	prSubrouter := r.PathPrefix("/pullRequest").Subrouter()

	teamDelivery.RegisterRoutes(teamSubrouter)
	userDelivery.RegisterRoutes(userSubrouter)
	pullRequestDelivery.RegisterRoutes(prSubrouter)

	return r
}
