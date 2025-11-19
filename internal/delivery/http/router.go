package http

import (
	"avito-assignment-2025-autumn/internal/delivery/http/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(team *handlers.TeamDelivery, user *handlers.UserDelivery, pr *handlers.PullRequestDelivery, middlewares ...mux.MiddlewareFunc) http.Handler {
	r := mux.NewRouter()

	for _, mw := range middlewares {
		if mw != nil {
			r.Use(mw)
		}
	}

	if team != nil {
		team.RegisterRoutes(r)
	}

	//if user != nil {
	//	user.RegisterRoutes(r)
	//}
	//if pr != nil {
	//	pr.RegisterRoutes(r)
	//}

	return r
}
