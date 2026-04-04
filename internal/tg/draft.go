package tg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gotd/td/tg"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/pkg/errors"
)

type DraftArguments struct {
	Name string `json:"name" jsonschema:"required,description=Name of the dialog"`
	Text string `json:"text" jsonschema:"required,description=Plain text of the message"`
}

type DraftResponse struct {
	Success bool `json:"success"`
}

func (c *Client) SendDraft(args DraftArguments) (*mcp.ToolResponse, error) {
	var ok bool
	client := c.T()
	if err := client.Run(context.Background(), func(ctx context.Context) (err error) {
		api := client.API()

		inputPeer, err := getInputPeerFromName(ctx, api, args.Name)
		if err != nil {
			return fmt.Errorf("get inputPeer from name: %w", err)
		}

		ok, err = api.MessagesSaveDraft(ctx, &tg.MessagesSaveDraftRequest{
			Peer:    inputPeer,
			Message: args.Text,
		})
		if err != nil {
			return fmt.Errorf("save draft: %w", err)
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to save draft")
	}

	jsonData, err := json.Marshal(DraftResponse{Success: ok})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response")
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(jsonData))), nil
}
