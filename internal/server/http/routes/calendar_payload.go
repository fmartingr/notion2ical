package routes

import (
	"fmt"
	"net/url"
	"strings"
)

type calendarDownloadPayload struct {
	DatabaseID   string `form:"database_id" query:"database_id"`
	AllDayEvents bool   `form:"all_day_events" query:"all_day_events"`
	DateProperty string `form:"date_property" query:"date_property"`
	NameProperty string `form:"name_property" query:"name_property"`
}

func (o calendarDownloadPayload) Validate() error {
	if o.DatabaseID == "" {
		return fmt.Errorf("Database ID can't be empty")
	}
	if o.DateProperty == "" {
		return fmt.Errorf("Date property can't be empty")
	}
	if o.NameProperty == "" {
		return fmt.Errorf("Name property can't be empty")
	}
	return nil
}

type wizardPayload struct {
	DatabaseURL string `form:"database_url"`

	databaseID string
}

func (o *wizardPayload) Validate() error {
	if o.DatabaseURL == "" {
		return fmt.Errorf("Notion database URL can't be empty")
	}

	databaseURL, err := url.Parse(o.DatabaseURL)
	if err != nil {
		return fmt.Errorf("Error parsing Notion database URL, is it correct?")
	}

	o.databaseID, err = o.parseDatabaseID(databaseURL)
	if err != nil {
		return err
	}

	return nil
}

func (o *wizardPayload) parseDatabaseID(databaseURL *url.URL) (databaseID string, err error) {
	if databaseURL == nil {
		return databaseID, fmt.Errorf("Notion database URL couldn't be parsed or is invalid")
	}

	parts := strings.Split(databaseURL.Path, "/")
	if len(parts) != 3 {
		return databaseID, fmt.Errorf("Notion database URL seems malformed")
	}

	return parts[2], nil
}

func (o *wizardPayload) GetDatabaseID() string {
	return o.databaseID
}
