package http

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/dstotijn/go-notion"
	"github.com/emersion/go-ical"
	notionClient "github.com/fmartingr/notion2ical/internal/notion"
	"github.com/fmartingr/notion2ical/internal/server/http/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/gofiber/fiber/v2/utils"
	"go.uber.org/zap"
)

type HttpServer struct {
	http   *fiber.App
	addr   string
	logger *zap.Logger
	notion *notionClient.NotionClient
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
		Get("/calendar/:databaseID", timeout.New(s.calendarHandler, time.Second*30)).
		Get("/liveness", s.livenessHandler).
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

func (s *HttpServer) livenessHandler(c *fiber.Ctx) error {
	return c.SendStatus(200)
}

type calendarOptions struct {
	AllDayEvents      bool   `query:"all_day_events"`
	DateFieldProperty string `query:"date_field_property"`
	NameProperty      string `query:"name_property"`
}

func (s *HttpServer) calendarHandler(c *fiber.Ctx) error {
	var options calendarOptions
	if err := c.QueryParser(&options); err != nil {
		s.logger.Error("error parsing query", zap.String("query", c.Context().QueryArgs().String()))
	}

	dbID := c.Params("databaseID")

	query := notion.DatabaseQuery{
		Filter: &notion.DatabaseQueryFilter{
			Property: options.DateFieldProperty,
			Date: &notion.DateDatabaseQueryFilter{
				IsNotEmpty: true,
			},
		},
	}

	result, err := s.notion.Client.QueryDatabase(c.Context(), dbID, &query)
	if err != nil {
		s.logger.Error("can't query notion database", zap.Error(err))
		return c.SendStatus(500)
	}

	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//notion2ical.fmartingr.dev//NONSGML PDA Calendar Version 1.0//EN")

	for _, r := range result.Results {
		event := ical.NewEvent()
		event.Props.SetText(ical.PropUID, r.ID)
		dateProperty := r.Properties.(notion.DatabasePageProperties)[options.DateFieldProperty].Date
		dateStart := dateProperty.Start.Time
		dateEnd := dateStart
		if dateProperty.End != nil {
			dateEnd = dateProperty.End.Time
		}
		if options.AllDayEvents {
			event.Props.SetDate(ical.PropDateTimeStamp, dateStart)
			event.Props.SetDate(ical.PropDateTimeStart, dateEnd)
		} else {
			event.Props.SetDateTime(ical.PropDateTimeStamp, dateStart)
			event.Props.SetDateTime(ical.PropDateTimeStart, dateEnd)
		}
		event.Props.SetText(ical.PropSummary, r.Properties.(notion.DatabasePageProperties)[options.NameProperty].Title[0].Text.Content)

		cal.Children = append(cal.Children, event.Component)
	}

	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(cal); err != nil {
		s.logger.Fatal("error encoding calendar", zap.Error(err))
	}

	return c.Send(buf.Bytes())
}

func NewHttpServer(logger *zap.Logger, port int, notionClient *notionClient.NotionClient) *HttpServer {
	server := HttpServer{
		logger: logger,
		addr:   fmt.Sprintf(":%d", port),
		notion: notionClient,
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
