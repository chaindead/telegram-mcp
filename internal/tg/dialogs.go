package tg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/td/tg"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// DialogType represents the type of dialog for filtering
type DialogType string

const (
	DialogTypeUnknown DialogType = "unknown"
	DialogTypeAll     DialogType = ""
	DialogTypeUser    DialogType = "user"
	DialogTypeBot     DialogType = "bot"
	DialogTypeChat    DialogType = "chat"
	DialogTypeChannel DialogType = "channel"

	DefaultDialogsLimit = 100
)

// nolint:lll
type DialogsArguments struct {
	Offset     string `json:"offset,omitempty" jsonschema:"description=Offset for continuation"`
	OnlyUnread bool   `json:"only_unread,omitempty" jsonschema:"description=Include only dialogs with unread mark"`
}

type MessageInfo struct {
	ID       int    `json:"id,omitempty"`
	Who      string `json:"who,omitempty"`
	When     string `json:"when"`
	Text     string `json:"text,omitempty"`
	IsUnread bool   `json:"is_unread,omitempty"`
	ts       int
}

type DialogInfo struct {
	Name        string       `json:"name,omitempty"`
	Type        string       `json:"type"`
	Title       string       `json:"title"`
	LastMessage *MessageInfo `json:"last_message,omitempty"`
	Empty       bool         `json:"empty,omitempty"`
}

type DialogsResponse struct {
	Dialogs []DialogInfo  `json:"dialogs"`
	Offset  DialogsOffset `json:"offset"`
}

// GetDialogs returns a list of dialogs (chats, channels, groups)
func (c *Client) GetDialogs(args DialogsArguments) (*mcp.ToolResponse, error) {
	var offset DialogsOffset
	if args.Offset != "" {
		if err := offset.UnmarshalJSON([]byte(args.Offset)); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal offset")
		}
	}
	if offset.Peer == nil {
		offset.Peer = &tg.InputPeerEmpty{}
	}

	var dc tg.MessagesDialogsClass
	client := c.T()
	if err := client.Run(context.Background(), func(ctx context.Context) (err error) {
		api := client.API()
		dc, err = api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: offset.Peer,
			OffsetID:   offset.MsgID,
			OffsetDate: offset.Date,
		})
		if err != nil {
			return fmt.Errorf("failed to get dialogs: %w", err)
		}

		// Debug
		//jsonData, _ := json.Marshal(dc)
		//log.Info().RawJSON("dialogs", cleanJSON(jsonData)).Msg("dialogs")

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to get dialogs")
	}

	d, err := newDialogs(dc, args.OnlyUnread)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dialogs")
	}

	rsp := DialogsResponse{
		Dialogs: d.Info(),
		Offset:  d.Offset(),
	}

	jsonData, err := json.Marshal(rsp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response")
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(jsonData))), nil
}

type dialogs struct {
	tg.MessagesDialogs

	// chat id key
	messages map[int64]*tg.Message
	users    map[int64]*tg.User
	//dialogs  map[string]*tg.Dialog
	chats    map[int64]*tg.Chat
	channels map[int64]*tg.Channel

	//opts
	onlyUnread bool
}

func newDialogs(rawD tg.MessagesDialogsClass, onlyUnread bool) (*dialogs, error) {
	var d dialogs
	switch dT := rawD.(type) {
	case *tg.MessagesDialogs:
		d = dialogs{MessagesDialogs: *dT}
	case *tg.MessagesDialogsSlice:
		d = dialogs{MessagesDialogs: tg.MessagesDialogs{
			Dialogs:  dT.Dialogs,
			Messages: dT.Messages,
			Chats:    dT.Chats,
			Users:    dT.Users,
		}}
	case *tg.MessagesDialogsNotModified:
	default:
	}

	d.messages = make(map[int64]*tg.Message)
	for _, m := range d.Messages {
		switch mT := m.(type) {
		case *tg.Message:
			d.messages[getPeerID(mT.PeerID)] = mT
		case *tg.MessageService, *tg.MessageEmpty:
		default:
		}
	}
	delete(d.messages, 0)

	d.users = make(map[int64]*tg.User)
	for _, uc := range d.Users {
		u, ok := uc.(*tg.User)
		if !ok {
			log.Debug().Msgf("newDialogs(%+v): invalid message user", uc)
			continue
		}

		d.users[u.GetID()] = u
	}

	d.chats = make(map[int64]*tg.Chat)
	d.channels = make(map[int64]*tg.Channel)
	for _, c := range d.Chats {
		switch cT := c.(type) {
		case *tg.Chat:
			d.chats[cT.ID] = cT
		case *tg.Channel:
			d.channels[cT.ID] = cT
		case *tg.ChatForbidden, *tg.ChannelForbidden, *tg.ChatEmpty:
		default:
		}
	}

	d.onlyUnread = onlyUnread

	return &d, nil
}

func (d *dialogs) Info() []DialogInfo {
	ds := make([]DialogInfo, 0, len(d.Dialogs))

	for _, dItem := range d.Dialogs {
		dialogItem, ok := dItem.(*tg.Dialog)
		if !ok {
			continue
		}

		if d.onlyUnread && dialogItem.UnreadCount == 0 {
			continue
		}

		info, err := d.processDialog(dialogItem)
		if err != nil {
			log.Debug().Err(err).Str("dialog", dItem.String()).Msg("failed process dialog")
			continue
		}

		if info.Title == "" {
			continue
		}

		ds = append(ds, info)
	}

	return ds
}

func (d *dialogs) Offset() DialogsOffset {
	for i := len(d.Dialogs) - 1; i >= 0; i-- {
		dialogItem, ok := d.Dialogs[i].(*tg.Dialog)
		if !ok {
			continue
		}
		if dialogItem.Peer == nil {
			continue
		}

		msg, ok := d.messages[getPeerID(dialogItem.Peer)]
		if !ok {
			continue
		}

		return DialogsOffset{
			MsgID: msg.ID,
			Date:  msg.Date,
			Peer:  getInputPeerID(dialogItem.Peer),
		}
	}

	return DialogsOffset{}
}

func (d *dialogs) processDialog(dialogItem *tg.Dialog) (DialogInfo, error) {
	var info DialogInfo

	if msg, ok := d.messages[getPeerID(dialogItem.Peer)]; ok {
		var who string
		if msg.FromID != nil {
			name, _, err := d.getNameID(msg.FromID)
			if err != nil {
				return DialogInfo{}, errors.Wrap(err, "failed to get dialog name")
			}
			who = name
		}

		// Limit message to 20 words
		text := msg.Message
		words := strings.Fields(text)
		if len(words) > 20 {
			text = strings.Join(words[:20], " ") + "..."
		}

		info.LastMessage = &MessageInfo{
			ID:       msg.ID,
			Who:      who,
			When:     time.Unix(int64(msg.Date), 0).Format(time.DateTime),
			ts:       msg.Date,
			Text:     text,
			IsUnread: dialogItem.UnreadCount > 0,
		}
	}

	if dialogItem.Peer == nil {
		return DialogInfo{}, fmt.Errorf("no peer: %s", dialogItem.String())
	}

	var err error
	info.Title, info.Name, err = d.getNameID(dialogItem.Peer)
	if err != nil {
		return DialogInfo{}, err
	}

	info.Type = string(d.getType(dialogItem))

	if info.LastMessage == nil {
		info.Empty = true
	}

	return info, nil
}

func (d *dialogs) getNameID(pC tg.PeerClass) (string, string, error) {
	var name, username string
	switch p := pC.(type) {
	case *tg.PeerUser:
		u, ok := d.users[p.GetUserID()]
		if !ok {
			return "", "", errors.Errorf("peerid(%d): invalid message user", p.GetUserID())
		}
		name = getTitle(u)
		username = getUsername(u)
	case *tg.PeerChannel:
		channel, ok := d.channels[p.GetChannelID()]
		if !ok {
			return "", "", errors.Errorf("peerid(%d): invalid message channel", p.GetChannelID())
		}

		name = getTitle(channel)
		username = getUsername(channel)
	case *tg.PeerChat:
		chat, ok := d.chats[p.GetChatID()]
		if !ok {
			return "", "", errors.Errorf("peerid(%d): invalid message chat", p.GetChatID())
		}

		name = getTitle(chat)
		username = getUsername(chat)
	default:
		return "", "", fmt.Errorf("chose author(%T): invalid dialog peer", p)
	}

	return name, username, nil
}

func (d *dialogs) getType(rawD *tg.Dialog) DialogType {
	switch v := rawD.Peer.(type) {
	case *tg.PeerChannel:
		return DialogTypeChannel
	case *tg.PeerChat:
		return DialogTypeChat
	case *tg.PeerUser:
		u, ok := d.users[getPeerID(rawD.Peer)]
		if !ok {
			log.Debug().Msgf("getType(%+v): user not found", v)
			return DialogTypeUser
		}

		if u.Bot {
			return DialogTypeBot
		}

		return DialogTypeUser
	default:
		log.Debug().Msgf("getType(%+v): unknown dialog type", v)
		return DialogTypeUnknown
	}
}

func getPeerID(p tg.PeerClass) int64 {
	if p == nil {
		return 0
	}

	switch v := p.(type) {
	case *tg.PeerChannel:
		return v.ChannelID
	case *tg.PeerChat:
		return v.ChatID
	case *tg.PeerUser:
		return v.UserID
	default:
		return 0
	}

}
