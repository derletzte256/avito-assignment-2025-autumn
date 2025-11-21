package http

import (
	"avito-assignment-2025-autumn/internal/config"
	"context"
	"errors"
	nethttp "net/http"

	"go.uber.org/zap"
)

type Server struct {
	httpServer *nethttp.Server
	logger     *zap.Logger
}

func NewServer(httpCfg config.HTTPConfig, handler nethttp.Handler, logger *zap.Logger) *Server {

	srv := &nethttp.Server{
		Addr:         httpCfg.Address,
		Handler:      handler,
		ReadTimeout:  httpCfg.ReadTimeout,
		WriteTimeout: httpCfg.WriteTimeout,
		IdleTimeout:  httpCfg.IdleTimeout,
	}

	return &Server{httpServer: srv, logger: logger}
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
