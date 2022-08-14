package http

import (
	"context"
	"fmt"
	"time"

	notionClient "github.com/fmartingr/notion2ical/internal/notion"
	"github.com/fmartingr/notion2ical/internal/server/http/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
	"go.uber.org/zap"
)

type HttpServer struct {
	http   *fiber.App
	addr   string
	logger *zap.Logger
	routes *routes
}

func (s *HttpServer) Setup() {
	s.http.
		Use(requestid.New(requestid.Config{
			Generator: utils.UUIDv4,
		})).
		Use(middleware.NewZapMiddleware(middleware.ZapMiddlewareConfig{
			Logger:      s.logger,
			CacheHeader: "X-Cache",
		})).
		Use(cache.New(cache.Config{
			Next: func(c *fiber.Ctx) bool {
				return c.Query("refresh") == "true"
			},
			Expiration:   60 * time.Minute,
			CacheControl: true,
			CacheHeader:  "X-Cache",
			// KeyGenerator: func(c *fiber.Ctx) string {
			// 	return utils.CopyString(c.Path() + string(c.Context().QueryArgs().QueryString()))
			// },
		})).
		Use(recover.New()).
		Static("/", "./public").
		Mount("/", s.routes.Router()).
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

func NewHttpServer(logger *zap.Logger, port int, notionClient *notionClient.NotionClient) *HttpServer {
	server := HttpServer{
		logger: logger,
		addr:   fmt.Sprintf(":%d", port),
		routes: newRoutes(logger, notionClient),
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
	server.Setup()

	return &server
}
