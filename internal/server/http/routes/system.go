package routes

import (
	"log/slog"
	"net/http"

	"github.com/fmartingr/notion2ical/internal/models"
)

func (a *API) handleLiveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (a *API) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"version": "` + models.BuildVersion + `"}`)); err != nil {
		a.logger.Error("error writing version response", slog.String("err", err.Error()))
	}
}
