package routes

import (
	"github.com/fmartingr/notion2ical/internal/config"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type SystemRoutes struct {
	logger *zap.Logger
	router *fiber.App
}

func (r *SystemRoutes) Setup() *SystemRoutes {
	r.router.
		Get("/liveness", r.livenessHandler)
	return r
}

func (r *SystemRoutes) Router() *fiber.App {
	return r.router
}

func (r *SystemRoutes) livenessHandler(c *fiber.Ctx) error {
	return c.SendStatus(200)
}

func NewSystemRoutes(logger *zap.Logger, _ *config.Config) *SystemRoutes {
	routes := SystemRoutes{
		logger: logger,
		router: fiber.New(),
	}
	routes.Setup()
	return &routes
}
