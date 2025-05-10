package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chaindead/telegram-mcp/internal/tg"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func serve(ctx context.Context, cmd *cli.Command) error {
	appID := cmd.Int("app-id")
	appHash := cmd.String("api-hash")
	sessionPath := cmd.String("session")
	dryRun := cmd.Bool("dry")

	_, err := os.Stat(sessionPath)
	if err != nil {
		return fmt.Errorf("session file not found(%s): %w", sessionPath, err)
	}

	server := mcp.NewServer(stdio.NewStdioServerTransport())
	client := tg.New(int(appID), appHash, sessionPath)

	if dryRun {
		answer, err := client.GetMe(tg.EmptyArguments{})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		data, err := json.MarshalIndent(answer, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		log.Info().RawJSON("answer", data).Msg("Check GetMe: OK")

		answer, err = client.GetDialogs(tg.DialogsArguments{Offset: "", OnlyUnread: true})
		if err != nil {
			return fmt.Errorf("get dialogs: %w", err)
		}

		log.Info().RawJSON("answer", []byte(answer.Content[0].TextContent.Text)).Msg("Check GetDialogs: OK")

		answer, err = client.GetHistory(tg.HistoryArguments{Name: os.Getenv("TG_TEST_USERNAME")})
		if err != nil {
			return fmt.Errorf("get nickname history: %w", err)
		}

		answer, err = client.GetHistory(tg.HistoryArguments{Name: "cht[4626931529]"})
		if err != nil {
			return fmt.Errorf("get chat history: %w", err)
		}

		answer, err = client.GetHistory(tg.HistoryArguments{Name: "chn[2225853048:8934705438195741763]"})
		if err != nil {
			return fmt.Errorf("get chan history: %w", err)
		}

		log.Info().RawJSON("answer", []byte(answer.Content[0].TextContent.Text)).Msg("Check GetHistory: OK")

		answer, err = client.SendDraft(tg.DraftArguments{Name: os.Getenv("TG_TEST_USERNAME"), Text: "test draft"})
		if err != nil {
			log.Err(err).Msg("Check SendDraft: FAIL")
		} else {
			log.Info().RawJSON("answer", []byte(answer.Content[0].TextContent.Text)).Msg("Check SendDraft: OK")
		}

		answer, err = client.ReadHistory(tg.ReadArguments{Name: os.Getenv("TG_TEST_USERNAME")})
		if err != nil {
			log.Err(err).Msg("Check ReadHistory: FAIL")
		} else {
			log.Info().RawJSON("answer", []byte(answer.Content[0].TextContent.Text)).Msg("Check ReadHistory: OK")
		}

		answer, err = client.GetUserMessages(tg.UserMessagesArguments{
			Name:     os.Getenv("TG_TEST_GROUPNAME"),
			Username: os.Getenv("TG_TEST_USERNAME"),
			Limit:    100,
		})
		if err != nil {
			log.Err(err).Msg("Check GetUserMessages: FAIL")
		} else {
			log.Info().RawJSON("answer", []byte(answer.Content[0].TextContent.Text)).Msg("Check GetUserMessages: OK")
		}

		return nil
	}

	err = server.RegisterTool("tg_me", "Get current telegram account info", client.GetMe)
	if err != nil {
		return fmt.Errorf("register tool: %w", err)
	}

	err = server.RegisterTool("tg_dialogs", "Get list of telegram dialogs (chats, channels, users)", client.GetDialogs)
	if err != nil {
		return fmt.Errorf("register dialogs tool: %w", err)
	}

	err = server.RegisterTool("tg_dialog", "Get messages of telegram dialog", client.GetHistory)
	if err != nil {
		return fmt.Errorf("register dialogs tool: %w", err)
	}

	err = server.RegisterTool("tg_send", "Send draft message to dialog", client.SendDraft)
	if err != nil {
		return fmt.Errorf("register dialogs tool: %w", err)
	}

	err = server.RegisterTool("tg_read", "Mark dialog messages as read", client.ReadHistory)
	if err != nil {
		return fmt.Errorf("register read tool: %w", err)
	}

	err = server.RegisterTool("tg_user_messages", "Get messages sent by a specific user in a group or channel", client.GetUserMessages)
	if err != nil {
		return fmt.Errorf("register user messages tool: %w", err)
	}

	if err := server.Serve(); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	<-ctx.Done()

	return nil
}
