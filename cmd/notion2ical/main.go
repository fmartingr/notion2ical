package main

import (
	"context"

	"github.com/fmartingr/notion2ical/internal/config"
	"github.com/fmartingr/notion2ical/internal/notion"
	"github.com/fmartingr/notion2ical/internal/server"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// TODO: set log level

	defer func() {
		if err := logger.Sync(); err != nil {
			panic(err)
		}
	}()

	cfg := config.ParseServerConfiguration(ctx, logger)

	cfg.Notion.Client = notion.NewNotionClient(logger, cfg.Notion.MaxPagination, cfg.Notion.IntegrationToken)

	server := server.NewServer(
		logger,
		cfg,
	)

	if err := server.Start(ctx); err != nil {
		logger.Panic("error starting server", zap.Error(err))
	}

	server.WaitStop()
}
