package main

import (
	"encoding/json"
	"fmt"

	mcp "github.com/metoro-io/mcp-golang"
)

func sampleResource() (*mcp.ResourceResponse, error) {
	type Chat struct {
		ID          int64  `json:"id,omitempty"`
		Type        string `json:"type"`
		Title       string `json:"title"`
		UnreadCount int    `json:"unread_count"`
	}

	chats := []Chat{
		{
			ID:          123456789,
			Type:        "channel",
			Title:       "Sample Channel",
			UnreadCount: 5,
		},
		{
			ID:          987654321,
			Type:        "group",
			Title:       "Test Group",
			UnreadCount: 2,
		},
	}

	rss := make([]*mcp.EmbeddedResource, 0, len(chats))
	for _, chat := range chats {
		chat.ID = 0
		uri := fmt.Sprintf("telegram://chats/%d", chat.ID)

		content, err := json.Marshal(chat)
		if err != nil {
			return nil, err
		}

		rss = append(rss, mcp.NewTextEmbeddedResource(uri, string(content), "application/json"))
	}

	return mcp.NewResourceResponse(rss...), nil
}
