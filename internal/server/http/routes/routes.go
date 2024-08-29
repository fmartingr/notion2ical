package routes

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/fmartingr/notion2ical/internal/config"
	notionClient "github.com/fmartingr/notion2ical/internal/notion"
	"github.com/fmartingr/notion2ical/internal/server/http/public"
	"github.com/fmartingr/notion2ical/internal/server/http/views"
)

type API struct {
	cfg    *config.Config
	logger *slog.Logger
	notion *notionClient.NotionClient
}

func (a *API) GetStaticPath(path string) string {
	return "/static/" + path
}

func (a *API) getTemplateFuncs() map[string]any {
	return map[string]any{
		"GetStaticPath": a.GetStaticPath,
	}
}

func (a *API) renderTemplate(templateName string, w http.ResponseWriter, data any) error {
	baseTmpl, err := template.New("base").Funcs(a.getTemplateFuncs()).ParseFS(views.Assets, "base.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("error parsing base template: %w", err)
	}

	indexTmpl, err := template.Must(baseTmpl.Clone()).Funcs(a.getTemplateFuncs()).ParseFS(views.Assets, templateName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("error parsing %s template: %w", templateName, err)
	}

	err = indexTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

func NewRouter(cfg *config.Config, logger *slog.Logger) *http.ServeMux {
	api := &API{
		cfg:    cfg,
		logger: logger,
		notion: cfg.Notion.Client,
	}
	fs := http.FileServer(http.FS(public.Assets))

	router := http.NewServeMux()

	router.HandleFunc("GET /system/liveness", api.handleLiveness)
	router.HandleFunc("GET /system/version", api.handleVersion)

	router.Handle("GET /static/", http.StripPrefix("/static", fs))
	router.HandleFunc("POST /download", api.downloadHandler)
	router.HandleFunc("GET /calendar.ics", api.calendarIcsHandler)
	router.HandleFunc("GET /", api.indexHandler)
	router.HandleFunc("POST /", api.indexHandler)

	router.HandleFunc("POST /wizard", api.wizardHandler)

	return router
}
