package notion

import (
	"context"
	"time"

	notion "github.com/dstotijn/go-notion"
	"go.uber.org/zap"
)

type NotionClient struct {
	Client *notion.Client
	logger *zap.Logger

	// maxPagination the maximum number of pages to retrieve items from the database
	maxPagination int
}

type NotionDatabaseInfo struct {
	ID             string
	Name           string
	TextProperties []string
	DateProperties []string
}

// GetDatabaseInfo returns datetime and text properties which can be used to format calendar events
func (c *NotionClient) GetDatabaseInfo(ctx context.Context, databaseID string) (info NotionDatabaseInfo, err error) {
	db, err := c.Client.FindDatabaseByID(ctx, databaseID)
	if err != nil {
		return
	}

	info.ID = db.ID
	info.Name = db.Title[0].Text.Content

	for name, prop := range db.Properties {
		if prop.Type == notion.DBPropTypeDate {
			info.DateProperties = append(info.DateProperties, name)
		} else if prop.Type == notion.DBPropTypeRichText || prop.Type == notion.DBPropTypeTitle {
			info.TextProperties = append(info.TextProperties, name)
		}
	}

	return
}

type NotionDatabaseItem struct {
	ID        string
	Name      string
	DateStart time.Time
	DateEnd   time.Time
}

func (c *NotionClient) queryDatabase(ctx context.Context, databaseID, nameProperty, dateProperty string, cursor *string) (notion.DatabaseQueryResponse, error) {
	query := notion.DatabaseQuery{
		Filter: &notion.DatabaseQueryFilter{
			And: []notion.DatabaseQueryFilter{
				{
					Property: dateProperty,
					Date: &notion.DateDatabaseQueryFilter{
						IsNotEmpty: true,
					},
				},
				{
					Property: nameProperty,
					Text: &notion.TextDatabaseQueryFilter{
						IsNotEmpty: true,
					},
				},
			},
		},
	}

	if cursor != nil {
		query.StartCursor = *cursor
	}

	return c.Client.QueryDatabase(ctx, databaseID, &query)
}

func (c *NotionClient) GetDatabaseItems(ctx context.Context, databaseID, nameProperty, dateProperty string) (items []NotionDatabaseItem, err error) {
	var currentCursor *string = nil
	currentPage := 1

	for currentPage <= c.maxPagination {
		var result notion.DatabaseQueryResponse
		result, err = c.queryDatabase(ctx, databaseID, nameProperty, dateProperty, currentCursor)
		if err != nil {
			c.logger.Error("can't query notion database", zap.Error(err))
			return
		}

		for _, r := range result.Results {
			var name string
			textProperty := r.Properties.(notion.DatabasePageProperties)[nameProperty]
			if textProperty.Type == notion.DBPropTypeTitle {
				name = r.Properties.(notion.DatabasePageProperties)[nameProperty].Title[0].Text.Content
			} else {
				name = r.Properties.(notion.DatabasePageProperties)[nameProperty].RichText[0].Text.Content
			}

			dateProperty := r.Properties.(notion.DatabasePageProperties)[dateProperty].Date

			dateEnd := dateProperty.Start.Time
			if dateProperty.End != nil {
				dateEnd = dateProperty.End.Time
			}

			items = append(items, NotionDatabaseItem{
				ID:        r.ID,
				Name:      name,
				DateStart: dateProperty.Start.Time,
				DateEnd:   dateEnd,
			})
		}

		if !result.HasMore {
			break
		} else {
			currentCursor = result.NextCursor
			currentPage++
		}
	}

	return
}

func NewNotionClient(logger *zap.Logger, maxPagination int, integrationToken string) *NotionClient {
	return &NotionClient{
		logger:        logger,
		Client:        notion.NewClient(integrationToken),
		maxPagination: maxPagination,
	}
}
