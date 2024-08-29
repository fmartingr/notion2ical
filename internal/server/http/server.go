package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/fmartingr/notion2ical/internal/config"
	"github.com/fmartingr/notion2ical/internal/server/http/routes"
)

type HttpServer struct {
	server *http.Server
	router *http.ServeMux
	logger *slog.Logger
}

func (s *HttpServer) Setup(cfg *config.Config) {
	s.router = routes.NewRouter(cfg, s.logger)
}

func (s *HttpServer) Start(_ context.Context) error {
	s.logger.Info("starting http server", slog.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *HttpServer) Stop(ctx context.Context) error {
	s.logger.Info("stoppping http server")
	return s.server.Shutdown(ctx)
}

func NewHttpServer(logger *slog.Logger, cfg *config.Config) *HttpServer {
	server := HttpServer{
		logger: logger,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Http.Port),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			Handler:      routes.NewRouter(cfg, logger),
		},
	}
	server.Setup(cfg)

	return &server
}
