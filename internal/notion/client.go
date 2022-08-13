package notion

import notion "github.com/dstotijn/go-notion"

type NotionClient struct {
	Client *notion.Client
}

func NewNotionClient(integrationToken string) *NotionClient {
	return &NotionClient{
		Client: notion.NewClient(integrationToken),
	}
}
