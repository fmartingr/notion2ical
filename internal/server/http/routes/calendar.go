package routes

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/emersion/go-ical"
)

func (a *API) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := struct {
		Error       string
		ExtraHeader string
	}{
		Error: r.URL.Query().Get("error"),
	}

	if err := a.renderTemplate("index.html.tmpl", w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *API) wizardHandler(w http.ResponseWriter, r *http.Request) {
	var payload wizardPayload

	if err := r.ParseForm(); err != nil {
		a.logger.Error("error parsing form", slog.String("form", r.Form.Encode()))
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}

	if err := payload.FromBodyForm(r.Form); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape(err.Error())+"#how-it-works", http.StatusTemporaryRedirect)
		return
	}

	if err := payload.Validate(); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape(err.Error())+"#how-it-works", http.StatusTemporaryRedirect)
		return
	}

	info, err := a.notion.GetDatabaseInfo(context.TODO(), payload.GetDatabaseID())
	if err != nil {
		a.logger.Error("error getting database inforamtion", slog.String("err", err.Error()))
		http.Redirect(w, r, "/?error=Error getting database information, have you set up the integration properly?#how-it-works", http.StatusTemporaryRedirect)
		return
	}

	if len(info.DateProperties) == 0 {
		http.Redirect(w, r, "/?error=Your database does not have any properties, at least one is required#how-it-works", http.StatusTemporaryRedirect)
		return
	}

	if len(info.TextProperties) == 0 {
		http.Redirect(w, r, "/?error=Your database does not have any properties, at least one is required#how-it-works", http.StatusTemporaryRedirect)
		return
	}

	data := struct {
		TextProperties     []string
		DatetimeProperties []string
		DatabaseName       string
		DatabaseID         string
	}{
		TextProperties:     info.TextProperties,
		DatetimeProperties: info.DateProperties,
		DatabaseName:       info.Name,
		DatabaseID:         info.ID,
	}

	if err := a.renderTemplate("wizard.html.tmpl", w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *API) downloadHandler(w http.ResponseWriter, r *http.Request) {
	var payload calendarDownloadPayload

	if err := r.ParseForm(); err != nil {
		a.logger.Error("error parsing form", slog.String("form", r.Form.Encode()))
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}

	if err := payload.FromBodyForm(r.Form); err != nil {
		http.Redirect(w, r, "/wizard?error="+url.QueryEscape(err.Error()), http.StatusTemporaryRedirect)
		return
	}

	if err := payload.Validate(); err != nil {
		http.Redirect(w, r, "/wizard?error="+url.QueryEscape(err.Error()), http.StatusTemporaryRedirect)
		return
	}

	data := struct {
		CalendarSubscriptionURL string
		CalendarICSURL          string
		CalendarCacheTime       string
	}{
		CalendarSubscriptionURL: a.cfg.Http.PublicHostname + "/calendar.ics?" + payload.ToURLValues().Encode(),
		CalendarICSURL:          a.cfg.Http.PublicHostname + "/calendar.ics?" + payload.ToURLValues().Encode(),
		CalendarCacheTime:       a.cfg.Routes.Calendar.CacheExpiration.String(),
	}

	if err := a.renderTemplate("download.html.tmpl", w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *API) calendarIcsHandler(w http.ResponseWriter, r *http.Request) {
	var payload calendarDownloadPayload

	if err := r.ParseForm(); err != nil {
		a.logger.Error("error parsing form", slog.String("form", r.Form.Encode()))
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}

	if err := payload.FromBodyForm(r.Form); err != nil {
		http.Redirect(w, r, "/wizard?error="+url.QueryEscape(err.Error()), http.StatusTemporaryRedirect)
		return
	}

	if err := payload.Validate(); err != nil {
		http.Redirect(w, r, "/wizard?error="+url.QueryEscape(err.Error()), http.StatusTemporaryRedirect)
		return
	}

	results, err := a.notion.GetDatabaseItems(context.TODO(), payload.DatabaseID, payload.NameProperty, payload.DateProperty)
	if err != nil {
		a.logger.Error("error getting database items", slog.String("err", err.Error()))
		http.Error(w, "error getting database items", http.StatusInternalServerError)
		return
	}

	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//notion2ical//NONSGML PDA Calendar Version 1.0//EN")

	uri, err := url.Parse(a.cfg.Http.PublicHostname + "/calendar.ics?" + payload.ToURLValues().Encode())
	if err != nil {
		a.logger.Error("error formatting calendar url", slog.String("err", err.Error()))
		http.Error(w, "error formatting calendar url", http.StatusInternalServerError)
		return
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
		a.logger.Error("error encoding calendar", slog.String("err", err.Error()))
		return
	}

	w.Header().Add("Content-Type", "text/calendar")
	w.Header().Add("Content-Disposition", "attachment; filename=calendar.ics")

	if _, err := w.Write(buf.Bytes()); err != nil {
		a.logger.Error("error writing calendar", slog.String("err", err.Error()))
	}
}
