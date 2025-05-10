package tg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gotd/td/tg"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/pkg/errors"
)

type UserMessagesArguments struct {
	Name     string `json:"name" jsonschema:"required,description=Name of the dialog (group or channel)"`
	Username string `json:"username" jsonschema:"required,description=Username of the user to filter messages for"`
	Offset   int    `json:"offset,omitempty" jsonschema:"description=Offset for continuation"`
	Limit    int    `json:"limit,omitempty" jsonschema:"description=Maximum number of messages to return"`
}

type UserMessagesResponse struct {
	Messages []MessageInfo `json:"messages"`
	Offset   int           `json:"offset,omitempty"`
}

// GetUserMessages returns messages sent by a specific user in a group or channel
func (c *Client) GetUserMessages(args UserMessagesArguments) (*mcp.ToolResponse, error) {
	var messagesClass tg.MessagesMessagesClass
	client := c.T()

	if args.Limit <= 0 {
		args.Limit = 100
	}

	if err := client.Run(context.Background(), func(ctx context.Context) (err error) {
		api := client.API()
		inputPeer, err := getInputPeerFromName(ctx, api, args.Name)
		if err != nil {
			return fmt.Errorf("get inputPeer from name: %w", err)
		}

		var fromPeer tg.InputPeerClass
		if args.Username != "" {
			resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
				Username: args.Username,
			})
			if err != nil {
				// If I can not resolve the username, fall back to client-side filtering
				messagesClass, err = api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
					Peer:     inputPeer,
					OffsetID: args.Offset,
					Limit:    args.Limit,
				})
				if err != nil {
					return fmt.Errorf("failed to get history: %w", err)
				}
				return nil
			}

			if len(resolved.Users) > 0 {
				if user, ok := resolved.Users[0].(*tg.User); ok {
					fromPeer = &tg.InputPeerUser{
						UserID:     user.ID,
						AccessHash: user.AccessHash,
					}
				}
			}

			if fromPeer == nil {
				messagesClass, err = api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
					Peer:     inputPeer,
					OffsetID: args.Offset,
					Limit:    args.Limit,
				})
				if err != nil {
					return fmt.Errorf("failed to get history: %w", err)
				}
				return nil
			}
		} else {
			// If no username is provided, just get all messages
			messagesClass, err = api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer:     inputPeer,
				OffsetID: args.Offset,
				Limit:    args.Limit,
			})
			if err != nil {
				return fmt.Errorf("failed to get history: %w", err)
			}
			return nil
		}

		if fromPeer != nil {
			messagesClass, err = api.MessagesSearch(ctx, &tg.MessagesSearchRequest{
				Peer:      inputPeer,
				Q:         "", // Empty query to match all messages
				FromID:    fromPeer,
				Filter:    &tg.InputMessagesFilterEmpty{},
				MinDate:   0,
				MaxDate:   0,
				OffsetID:  args.Offset,
				AddOffset: 0,
				Limit:     args.Limit,
				MaxID:     0,
				MinID:     0,
				Hash:      0,
			})
			if err != nil {
				return fmt.Errorf("failed to search messages: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to get messages")
	}

	h, err := newHistory(messagesClass)
	if err != nil {
		return nil, errors.Wrap(err, "failed to process history")
	}

	var messages []MessageInfo
	if h != nil {
		messages = h.Info()
		if messages != nil && args.Username != "" {
			messages = filterMessagesByUsername(messages, args.Username)
		}
	}

	rsp := UserMessagesResponse{
		Messages: messages,
		Offset:   h.Offset(),
	}

	jsonData, err := json.Marshal(rsp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response")
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(jsonData))), nil
}

// filterMessagesByUsername filters messages by the sender's username
func filterMessagesByUsername(messages []MessageInfo, username string) []MessageInfo {
	var filtered []MessageInfo

	for _, msg := range messages {
		if msg.Who == username {
			filtered = append(filtered, msg)
		}
	}

	return filtered
}
