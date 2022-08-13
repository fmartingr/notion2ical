package server

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type httpServer struct {
	http *fiber.App
	addr string

	logger *zap.Logger
}

func (s *httpServer) Start(_ context.Context) error {
	s.http.
		Static("/", "./public").
		Get("/calendar", s.calendarHandler).
		Get("/liveness", s.livenessHandler).
		Use(s.notFound)

	s.logger.Info("starting http server", zap.String("addr", s.addr))
	return s.http.Listen(s.addr)
}

func (s *httpServer) Stop(ctx context.Context) error {
	s.logger.Info("stoppping http server")
	return s.http.Shutdown()
}

func (s *httpServer) notFound(c *fiber.Ctx) error {
	return c.SendStatus(404)
}

func (s *httpServer) livenessHandler(c *fiber.Ctx) error {
	return c.SendStatus(200)
}

func (s *httpServer) calendarHandler(c *fiber.Ctx) error {
	return c.SendStatus(501)
}

func NewHttpServer(logger *zap.Logger, port int) *httpServer {
	return &httpServer{
		logger: logger,
		addr:   fmt.Sprintf(":%d", port),
		http: fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				logger.Error(
					"handler error",
					zap.String("method", c.Method()),
					zap.String("path", c.Path()),
					zap.Error(err),
				)
				return c.SendStatus(500)
			},
		}),
	}
}
