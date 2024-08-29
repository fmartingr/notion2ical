package main

import (
	"context"
	"log/slog"

	"github.com/fmartingr/notion2ical/internal/config"
	"github.com/fmartingr/notion2ical/internal/models"
	"github.com/fmartingr/notion2ical/internal/notion"
	"github.com/fmartingr/notion2ical/internal/server"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	models.BuildVersion = version
	models.BuildCommit = commit
	models.BuildDate = date
}

func main() {
	ctx := context.Background()
	logger := slog.Default()

	cfg, err := config.ParseServerConfiguration(ctx, logger)
	if err != nil {
		logger.Error("error parsing configuration", slog.String("err", err.Error()))
		return
	}

	cfg.Notion.Client = notion.NewNotionClient(logger, cfg.Notion.MaxPagination, cfg.Notion.IntegrationToken)

	server := server.NewServer(
		logger,
		cfg,
	)

	if err := server.Start(ctx); err != nil {
		logger.Error("error starting server", slog.String("err", err.Error()))
	}

	server.WaitStop()
}
