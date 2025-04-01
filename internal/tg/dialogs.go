package tg

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/gotd/td/tg"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/pkg/errors"
)

// DialogType represents the type of dialog for filtering
type DialogType string

const (
	// DialogTypeAll represents all types of dialogs
	DialogTypeAll DialogType = ""
	// DialogTypeUser represents user chats
	DialogTypeUser DialogType = "user"
	// DialogTypeChat represents group chats
	DialogTypeChat DialogType = "chat"
	// DialogTypeChannel represents channels
	DialogTypeChannel DialogType = "channel"

	// DefaultDialogsLimit is the default limit for dialogs
	DefaultDialogsLimit = 100
)

// DialogsArguments contains parameters for getting dialogs
type DialogsArguments struct {
	Type  DialogType `json:"type,omitempty" jsonschema:"description=Filter dialogs by type (user, chat, channel or empty for all),enum=,enum=user,enum=chat,enum=channel"`
	Limit int        `json:"limit,omitempty" jsonschema:"description=Maximum number of dialogs to return (max: 100),default=100"`
}

// DialogInfo represents a simplified dialog structure
type DialogInfo struct {
	ID            int64  `json:"id"`
	Type          string `json:"type"`
	Title         string `json:"title"`
	UnreadCount   int    `json:"unread_count"`
	LastMessageID int    `json:"last_message_id"`
	IsVerified    bool   `json:"is_verified,omitempty"`
}

// GetDialogs returns a list of dialogs (chats, channels, groups)
func (c *Client) GetDialogs(args DialogsArguments) (*mcp.ToolResponse, error) {
	var result []DialogInfo

	if args.Limit <= 0 || args.Limit > DefaultDialogsLimit {
		args.Limit = DefaultDialogsLimit
	}

	if args.Type == "" {
		args.Type = DialogTypeAll
	}

	client := c.T()
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		api := client.API()
		dialogsClass, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: &tg.InputPeerEmpty{},
		})
		if err != nil {
			return fmt.Errorf("failed to get dialogs: %w", err)
		}

		var dialogs *tg.MessagesDialogs
		switch d := dialogsClass.(type) {
		case *tg.MessagesDialogs:
			dialogs = d
		case *tg.MessagesDialogsSlice:
			dialogs = &tg.MessagesDialogs{
				Dialogs:  d.Dialogs,
				Messages: d.Messages,
				Chats:    d.Chats,
				Users:    d.Users,
			}
		default:
			return fmt.Errorf("unexpected dialogs response type")
		}

		result = make([]DialogInfo, 0, len(dialogs.Dialogs))

		for _, dialog := range dialogs.Dialogs {
			dialogItem, ok := dialog.(*tg.Dialog)
			if !ok {
				continue
			}

			var info DialogInfo
			info.UnreadCount = dialogItem.UnreadCount
			info.LastMessageID = dialogItem.TopMessage

			switch peer := dialogItem.Peer.(type) {
			case *tg.PeerUser:
				if args.Type != DialogTypeAll && args.Type != DialogTypeUser {
					continue
				}

				for _, userItem := range dialogs.Users {
					user, ok := userItem.(*tg.User)
					if !ok || user.ID != peer.UserID {
						continue
					}

					info.ID = user.ID
					info.Type = "user"
					info.Title = getUserName(user)
					info.IsVerified = user.Verified

					result = append(result, info)
					break
				}

			case *tg.PeerChat:
				if args.Type != DialogTypeAll && args.Type != DialogTypeChat {
					continue
				}

				for _, chatItem := range dialogs.Chats {
					chat, ok := chatItem.(*tg.Chat)
					if !ok || chat.ID != peer.ChatID {
						continue
					}

					info.ID = chat.ID
					info.Type = "chat"
					info.Title = chat.Title

					result = append(result, info)
					break
				}

			case *tg.PeerChannel:
				if args.Type != DialogTypeAll && args.Type != DialogTypeChannel {
					continue
				}

				for _, channelItem := range dialogs.Chats {
					channel, ok := channelItem.(*tg.Channel)
					if !ok || channel.ID != peer.ChannelID {
						continue
					}

					info.ID = channel.ID
					info.Type = "channel"
					info.Title = channel.Title
					info.IsVerified = channel.Verified

					result = append(result, info)
					break
				}
			}
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to get dialogs")
	}

	// Convert response to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response")
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].LastMessageID > result[j].LastMessageID
	})

	if len(result) > args.Limit {
		result = result[:args.Limit]
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(jsonData))), nil
}

// Helper function to get user's name
func getUserName(user *tg.User) string {
	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	return name
}
