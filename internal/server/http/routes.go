package http

import (
	"bytes"

	"github.com/dstotijn/go-notion"
	"github.com/emersion/go-ical"
	notionClient "github.com/fmartingr/notion2ical/internal/notion"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type routes struct {
	logger *zap.Logger
	router *fiber.App
	notion *notionClient.NotionClient
}

func (r *routes) Setup() *routes {
	r.router.
		Get("/liveness", r.livenessHandler).
		Get("/calendar/:databaseID/", r.calendarHandler)
	return r
}

func (r *routes) Router() *fiber.App {
	return r.router
}

func (r *routes) livenessHandler(c *fiber.Ctx) error {
	return c.SendStatus(200)
}

type calendarOptions struct {
	AllDayEvents      bool   `query:"all_day_events"`
	DateFieldProperty string `query:"date_field_property"`
	NameProperty      string `query:"name_property"`
}

func (r *routes) calendarHandler(c *fiber.Ctx) error {
	var options calendarOptions
	if err := c.QueryParser(&options); err != nil {
		r.logger.Error("error parsing query", zap.String("query", c.Context().QueryArgs().String()))
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

	result, err := r.notion.Client.QueryDatabase(c.Context(), dbID, &query)
	if err != nil {
		r.logger.Error("can't query notion database", zap.Error(err))
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
		r.logger.Fatal("error encoding calendar", zap.Error(err))
	}

	return c.Send(buf.Bytes())
}

func newRoutes(logger *zap.Logger, notion *notionClient.NotionClient) *routes {
	routes := routes{
		logger: logger,
		notion: notion,
		router: fiber.New(),
	}
	routes.Setup()
	return &routes
}
