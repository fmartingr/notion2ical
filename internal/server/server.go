package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	internalModels "github.com/fmartingr/notion2ical/internal/models"
	"go.uber.org/zap"
)

type Server struct {
	Http   internalModels.Server
	config *ServerConfig
	logger *zap.Logger

	cancel context.CancelFunc
}

func (s *Server) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	if s.config.Http.Enabled {
		go func() {
			if err := s.Http.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.logger.Fatal("error starting server", zap.Error(err))
			}
		}()
	}

	return nil
}

func (s *Server) WaitStop() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signals
	s.logger.Info("signal received, shutting down", zap.String("signal", sig.String()))

	s.Stop()
}

func (s *Server) Stop() {
	s.cancel()

	shuwdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if s.config.Http.Enabled {
		if err := s.Http.Stop(shuwdownContext); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("error shutting down http server", zap.Error(err))
		}
	}
}

func NewServer(logger *zap.Logger, conf *ServerConfig) *Server {
	server := &Server{
		logger: logger,
		config: conf,
	}
	if conf.Http.Enabled {
		server.Http = NewHttpServer(logger, conf.Http.Port)
	}

	return server
}
