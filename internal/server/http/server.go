package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fmartingr/notion2ical/internal/config"
	"github.com/fmartingr/notion2ical/internal/server/http/middleware"
	"github.com/fmartingr/notion2ical/internal/server/http/routes"
	"github.com/fmartingr/notion2ical/internal/server/http/views"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/template/django"
	"go.uber.org/zap"
)

type HttpServer struct {
	http   *fiber.App
	addr   string
	logger *zap.Logger
}

func (s *HttpServer) Setup(cfg *config.Config) {
	s.http.
		Use(requestid.New(requestid.Config{
			Generator: utils.UUIDv4,
		})).
		Use(middleware.NewZapMiddleware(middleware.ZapMiddlewareConfig{
			Logger:      s.logger,
			CacheHeader: "X-Cache",
		})).
		Use(func(c *fiber.Ctx) error {
			c.Locals("branding_thanks_message", cfg.Branding.ThanksMessage)
			c.Locals("branding_footer_extra", cfg.Branding.FooterExtraMessage)
			c.Locals("calendar_cache_time", cfg.Routes.Calendar.CacheExpiration.String())
			return c.Next()
		}).
		Use(recover.New()).
		Mount(cfg.Routes.System.Path, routes.NewSystemRoutes(s.logger, cfg).Setup().Router()).
		Mount(cfg.Routes.Static.Path, routes.NewStaticRoutes(s.logger, cfg).Setup().Router()).
		Mount("/", routes.NewCalendarRoutes(s.logger, cfg).Setup().Router()).
		Use(s.notFound)
}

func (s *HttpServer) Start(_ context.Context) error {
	s.logger.Info("starting http server", zap.String("addr", s.addr))
	return s.http.Listen(s.addr)
}

func (s *HttpServer) Stop(ctx context.Context) error {
	s.logger.Info("stoppping http server")
	return s.http.Shutdown()
}

func (s *HttpServer) notFound(c *fiber.Ctx) error {
	return c.SendStatus(404)
}

func NewHttpServer(logger *zap.Logger, cfg *config.Config) *HttpServer {
	engine := django.NewFileSystem(http.FS(views.Assets), ".django")
	engine.AddFunc("static", func(path interface{}) string {
		if s, ok := path.(string); ok {
			return cfg.Routes.Static.Path + s
		}
		return cfg.Routes.Static.Path
	})
	server := HttpServer{
		logger: logger,
		addr:   fmt.Sprintf(":%d", cfg.Http.Port),
		http: fiber.New(fiber.Config{
			AppName:                      "notion2ical",
			PassLocalsToViews:            true,
			Views:                        engine,
			BodyLimit:                    cfg.Http.BodyLimit,
			ReadTimeout:                  cfg.Http.ReadTimeout,
			WriteTimeout:                 cfg.Http.WriteTimeout,
			IdleTimeout:                  cfg.Http.IDLETimeout,
			DisableKeepalive:             cfg.Http.DisableKeepAlive,
			DisablePreParseMultipartForm: cfg.Http.DisablePreParseMultipartForm,
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
	server.Setup(cfg)

	return &server
}
