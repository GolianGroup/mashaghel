package server

import (
	"context"
	"mashaghel/internal/config"
	"net/http"

	"go.uber.org/zap"
)

type HttpServer interface {
	Start() error
	Stop(ctx context.Context) error
	Mux() *http.ServeMux
}

type httpServer struct {
	mux    *http.ServeMux
	server *http.Server
	logger *zap.Logger
}

func NewHttpServer(cfg config.Config, logger *zap.Logger) HttpServer {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    config.SystemServerAddr(&cfg.Server),
		Handler: mux,
	}

	return &httpServer{
		mux:    mux,
		server: srv,
		logger: logger,
	}
}

func (s *httpServer) Start() error {
	err := s.server.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			s.logger.Error("Http server closed error", zap.Error(err))
			return err
		}
		s.logger.Error("Failed to start http server")
		return err
	}

	return nil
}

func (s *httpServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *httpServer) Mux() *http.ServeMux {
	return s.mux
}
