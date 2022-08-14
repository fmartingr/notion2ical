package routes

import (
	"net/http"
	"time"

	"github.com/fmartingr/notion2ical/internal/config"
	"github.com/fmartingr/notion2ical/internal/server/http/public"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"go.uber.org/zap"
)

type StaticRoutes struct {
	logger *zap.Logger
	router *fiber.App
	maxAge time.Duration
}

func (r *StaticRoutes) Setup() *StaticRoutes {
	r.router.
		Use(compress.New()).
		Use("/css", filesystem.New(filesystem.Config{
			Browse:     false,
			MaxAge:     int(r.maxAge.Seconds()),
			PathPrefix: "css",
			Root:       http.FS(public.Assets),
		})).
		Use("/images", filesystem.New(filesystem.Config{
			Browse:     false,
			MaxAge:     int(r.maxAge.Seconds()),
			PathPrefix: "images",
			Root:       http.FS(public.Assets),
		}))
	return r
}

func (r *StaticRoutes) Router() *fiber.App {
	return r.router
}

func NewStaticRoutes(logger *zap.Logger, cfg *config.Config) *StaticRoutes {
	routes := StaticRoutes{
		logger: logger,
		router: fiber.New(),
		maxAge: cfg.Routes.Static.MaxAge,
	}
	routes.Setup()
	return &routes
}
