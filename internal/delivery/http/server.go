package http

import (
	"avito-assignment-2025-autumn/internal/config"
	"avito-assignment-2025-autumn/internal/delivery/http/handlers"
	"context"
	"errors"
	nethttp "net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	httpServer *nethttp.Server
}

func NewServer(httpCfg config.HTTPConfig, team *handlers.TeamDelivery, user *handlers.UserDelivery, pr *handlers.PullRequestDelivery, middlewares ...mux.MiddlewareFunc) *Server {
	handler := NewRouter(team, user, pr, middlewares...)

	srv := &nethttp.Server{
		Addr:         httpCfg.Address,
		Handler:      handler,
		ReadTimeout:  httpCfg.ReadTimeout,
		WriteTimeout: httpCfg.WriteTimeout,
		IdleTimeout:  httpCfg.IdleTimeout,
	}

	return &Server{httpServer: srv}
}

func (s *Server) Start() error {
	if s == nil || s.httpServer == nil {
		return errors.New("http server is not configured")
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.httpServer == nil {
		return errors.New("http server is not configured")
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Addr() string {
	if s == nil || s.httpServer == nil {
		return ""
	}
	return s.httpServer.Addr
}
