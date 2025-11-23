package http

import (
	"net/http"

	prdelivery "github.com/derletzte256/avito-assignment-2025-autumn/internal/delivery/http/handlers/pullRequest"
	teamdelivery "github.com/derletzte256/avito-assignment-2025-autumn/internal/delivery/http/handlers/team"
	userdelivery "github.com/derletzte256/avito-assignment-2025-autumn/internal/delivery/http/handlers/user"
	"github.com/derletzte256/avito-assignment-2025-autumn/internal/delivery/http/middleware/accesslog"
	prrepo "github.com/derletzte256/avito-assignment-2025-autumn/internal/repo/postgres/pullRequest"
	teamrepo "github.com/derletzte256/avito-assignment-2025-autumn/internal/repo/postgres/team"
	userrepo "github.com/derletzte256/avito-assignment-2025-autumn/internal/repo/postgres/user"
	prusecase "github.com/derletzte256/avito-assignment-2025-autumn/internal/usecase/pullRequest"
	teamusecase "github.com/derletzte256/avito-assignment-2025-autumn/internal/usecase/team"
	userusecase "github.com/derletzte256/avito-assignment-2025-autumn/internal/usecase/user"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(pool *pgxpool.Pool, trManager *manager.Manager) http.Handler {
	r := mux.NewRouter()

	r.Use(accesslog.Middleware())

	teamRepo := teamrepo.NewTeamRepo(pool, trmpgx.DefaultCtxGetter)
	userRepo := userrepo.NewUserRepo(pool, trmpgx.DefaultCtxGetter)
	pullRequestRepo := prrepo.NewRepo(pool, trmpgx.DefaultCtxGetter)

	teamUC := teamusecase.NewUseCase(teamRepo, userRepo, trManager)
	teamDelivery := teamdelivery.NewTeamDelivery(teamUC)

	userUC := userusecase.NewUseCase(userRepo, pullRequestRepo, teamRepo, trManager)
	userDelivery := userdelivery.NewUserDelivery(userUC)

	pullRequestUC := prusecase.NewUseCase(pullRequestRepo, userRepo, teamRepo, trManager)
	pullRequestDelivery := prdelivery.NewDelivery(pullRequestUC)

	teamDelivery.RegisterRoutes(r)
	userDelivery.RegisterRoutes(r)
	pullRequestDelivery.RegisterRoutes(r)

	return r
}
