package tg

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/gotd/td/tg"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/pkg/errors"
)

type SendArguments struct {
	Name    string `json:"name" jsonschema:"required,description=Name of the dialog"`
	Text    string `json:"text" jsonschema:"required,description=Plain text of the message"`
	ReplyTo int    `json:"reply_to,omitempty" jsonschema:"description=Message ID to reply to"`
}

type SendResponse struct {
	Success bool `json:"success"`
}

func (c *Client) SendMessage(args SendArguments) (*mcp.ToolResponse, error) {
	var success bool
	client := c.T()

	if err := client.Run(context.Background(), func(ctx context.Context) (err error) {
		api := client.API()

		inputPeer, err := getInputPeerFromName(ctx, api, args.Name)
		if err != nil {
			return fmt.Errorf("get inputPeer from name: %w", err)
		}

		n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
		if err != nil {
			return fmt.Errorf("generate random ID: %w", err)
		}

		req := &tg.MessagesSendMessageRequest{
			Peer:     inputPeer,
			Message:  args.Text,
			RandomID: n.Int64(),
		}

		if args.ReplyTo > 0 {
			req.ReplyTo = &tg.InputReplyToMessage{
				ReplyToMsgID: args.ReplyTo,
			}
		}

		_, err = api.MessagesSendMessage(ctx, req)
		if err != nil {
			return fmt.Errorf("send message: %w", err)
		}

		success = true
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to send message")
	}

	jsonData, err := json.Marshal(SendResponse{Success: success})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response")
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(jsonData))), nil
}
