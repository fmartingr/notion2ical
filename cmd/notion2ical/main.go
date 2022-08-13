package main

import (
	"context"

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

	config := server.ParseServerConfiguration(ctx, logger)

	notionClient := notion.NewNotionClient(config.Notion.IntegrationToken)

	server := server.NewServer(
		logger,
		config,
		notionClient,
	)

	if err := server.Start(ctx); err != nil {
		logger.Panic("error starting server", zap.Error(err))
	}

	server.WaitStop()
}
