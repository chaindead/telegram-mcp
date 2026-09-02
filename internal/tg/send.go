package tg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gotd/td/telegram/message"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/pkg/errors"
)

type SendArguments struct {
	Name string `json:"name" jsonschema:"required,description=Name of the dialog"`
	Text string `json:"text" jsonschema:"required,description=Plain text of the message"`
}

type SendResponse struct {
	Success bool `json:"success"`
}

func (c *Client) SendMessage(args SendArguments) (*mcp.ToolResponse, error) {
	client := c.T()
	if err := client.Run(context.Background(), func(ctx context.Context) (err error) {
		api := client.API()

		inputPeer, err := getInputPeerFromName(ctx, api, args.Name)
		if err != nil {
			return fmt.Errorf("get inputPeer from name: %w", err)
		}

		sender := message.NewSender(api)
		if _, err = sender.To(inputPeer).Text(ctx, args.Text); err != nil {
			return fmt.Errorf("send message: %w", err)
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to send message")
	}

	jsonData, err := json.Marshal(SendResponse{Success: true})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response")
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(jsonData))), nil
}
