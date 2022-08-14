package routes

import (
	"bytes"
	"net/url"
	"time"

	"github.com/emersion/go-ical"
	"github.com/fmartingr/notion2ical/internal/config"
	notionClient "github.com/fmartingr/notion2ical/internal/notion"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/utils"
	"go.uber.org/zap"
)

type CalendarRoutes struct {
	logger         *zap.Logger
	router         *fiber.App
	notion         *notionClient.NotionClient
	publicHostname string
	thanksMessage  string
}

func (r *CalendarRoutes) Setup() *CalendarRoutes {
	r.router.
		Use(limiter.New(limiter.Config{
			Max:        2,
			Expiration: time.Second,
			// LimitReached: func(c *fiber.Ctx) error {
			// 	return c.SendFile("./toofast.html")
			// },
		})).
		Post("/wizard", r.wizardHandler).
		Post("/download", r.downloadHandler).
		Get("/", r.indexHandler).
		Use(cache.New(cache.Config{
			Expiration:   24 * time.Hour,
			CacheControl: true,
			CacheHeader:  "X-Cache",
			KeyGenerator: func(c *fiber.Ctx) string {
				return utils.CopyString(c.Path() + string(c.Context().QueryArgs().QueryString()))
			},
		})).
		Get("/calendar.ics", r.calendarIcsHandler)
	return r
}

func (r *CalendarRoutes) Router() *fiber.App {
	return r.router
}

func (r *CalendarRoutes) indexHandler(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"error": c.Query("error"),
	})
}

func (r *CalendarRoutes) wizardHandler(c *fiber.Ctx) error {
	var payload wizardPayload
	if err := c.BodyParser(&payload); err != nil {
		r.logger.Error("error parsing query", zap.String("query", c.Context().QueryArgs().String()))
		return err
	}

	if err := payload.Validate(); err != nil {
		return c.Redirect("/?error="+err.Error()+"#how-it-works", fiber.StatusTemporaryRedirect)
	}

	info, err := r.notion.GetDatabaseInfo(c.Context(), payload.GetDatabaseID())
	if err != nil {
		return c.Redirect("/?error=Error getting database information, have you set up the integration properly?#how-it-works", fiber.StatusTemporaryRedirect)
	}

	if len(info.DateProperties) == 0 {
		return c.Redirect("/?error=Your database does not have any datetime properties, at least one is required#how-it-works", fiber.StatusTemporaryRedirect)
	}

	if len(info.TextProperties) == 0 {
		return c.Redirect("/?error=Your database does not have any text properties, at least one is required#how-it-works", fiber.StatusTemporaryRedirect)
	}

	return c.Render("wizard", fiber.Map{
		"textProperties":     info.TextProperties,
		"datetimeProperties": info.DateProperties,
		"databaseName":       info.Name,
		"databaseID":         info.ID,
	})
}

func (r *CalendarRoutes) downloadHandler(c *fiber.Ctx) error {
	var payload calendarDownloadPayload
	if err := c.BodyParser(&payload); err != nil {
		r.logger.Error("error parsing query", zap.String("query", c.Context().QueryArgs().String()))
		return err
	}

	if err := payload.Validate(); err != nil {
		return c.Redirect("/wizard?error="+err.Error(), fiber.StatusTemporaryRedirect)
	}

	return c.Render("download", fiber.Map{
		"thanksMessage":           r.thanksMessage,
		"calendarSubscriptionUrl": r.publicHostname + "/calendar.ics?" + string(c.Request().Body()),
		"calendarICSUrl":          r.publicHostname + "/calendar.ics?" + string(c.Request().Body()),
	})
}

func (r *CalendarRoutes) calendarIcsHandler(c *fiber.Ctx) error {
	var payload calendarDownloadPayload
	if err := c.QueryParser(&payload); err != nil {
		r.logger.Error("error parsing query", zap.String("query", c.Context().QueryArgs().String()))
		return err
	}

	if err := payload.Validate(); err != nil {
		return c.Redirect("/?error="+err.Error(), fiber.StatusTemporaryRedirect)
	}

	results, err := r.notion.GetDatabaseItems(c.Context(), payload.DatabaseID, payload.NameProperty, payload.DateProperty)
	if err != nil {
		return err
	}

	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//notion2ical//NONSGML PDA Calendar Version 1.0//EN")

	uri, err := url.Parse(r.publicHostname + c.OriginalURL())
	if err != nil {
		r.logger.Error("error formatting calendar url", zap.Error(err))
		return err
	}
	cal.Props.SetURI(ical.PropURL, uri)

	for _, item := range results {
		event := ical.NewEvent()
		event.Props.SetText(ical.PropUID, item.ID)

		if payload.AllDayEvents {
			event.Props.SetDate(ical.PropDateTimeStamp, item.DateStart)
			event.Props.SetDate(ical.PropDateTimeStart, item.DateEnd)
		} else {
			event.Props.SetDateTime(ical.PropDateTimeStamp, item.DateStart)
			event.Props.SetDateTime(ical.PropDateTimeStart, item.DateEnd)
		}

		event.Props.SetText(ical.PropSummary, item.Name)

		cal.Children = append(cal.Children, event.Component)
	}

	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(cal); err != nil {
		r.logger.Error("error encoding calendar", zap.Error(err))
		return err
	}

	// c.Set("Content-Type", "text/calendar")
	return c.Send(buf.Bytes())
}

func NewCalendarRoutes(logger *zap.Logger, cfg *config.Config) *CalendarRoutes {
	routes := CalendarRoutes{
		logger:         logger,
		notion:         cfg.Notion.Client,
		router:         fiber.New(),
		publicHostname: cfg.Http.PublicHostname,
		thanksMessage:  cfg.Branding.ThanksMessage,
	}
	routes.Setup()
	return &routes
}
