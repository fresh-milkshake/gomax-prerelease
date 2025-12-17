package main

import "time"

type MessageEventRequest struct {
	ChatID    int64  `json:"chatId"`
	MessageID int64  `json:"messageId"`
	Text      string `json:"text"`
	SenderID  int64  `json:"senderId"`
	Time      int64  `json:"time,omitempty"`
}

type MessageEditEventRequest struct {
	ChatID    int64  `json:"chatId"`
	MessageID int64  `json:"messageId"`
	Text      string `json:"text"`
	SenderID  int64  `json:"senderId"`
	Time      int64  `json:"time,omitempty"`
}

type MessageDeleteEventRequest struct {
	ChatID     int64   `json:"chatId"`
	MessageID  int64   `json:"messageId"`
	MessageIDs []int64 `json:"messageIds,omitempty"`
}

type ReactionEventRequest struct {
	ChatID       int64            `json:"chatId"`
	MessageID    string           `json:"messageId"`
	Reaction     string           `json:"reaction,omitempty"`
	TotalCount   int              `json:"totalCount,omitempty"`
	YourReaction string           `json:"yourReaction,omitempty"`
	Counters     []map[string]any `json:"counters,omitempty"`
}

type ChatUpdateEventRequest struct {
	ChatID int64          `json:"chatId"`
	Chat   map[string]any `json:"chat,omitempty"`
	Data   map[string]any `json:"data,omitempty"`
}

type ContactUpdateEventRequest struct {
	Contact map[string]any `json:"contact,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

type TypingEventRequest struct {
	ChatID int64 `json:"chatId"`
	UserID int64 `json:"userId"`
	Typing bool  `json:"typing"`
}

func BuildMessageEvent(req MessageEventRequest) map[string]any {
	if req.Time == 0 {
		req.Time = time.Now().UnixMilli()
	}

	message := map[string]any{
		"id":       req.MessageID,
		"chatId":   req.ChatID,
		"text":     req.Text,
		"senderId": req.SenderID,
		"time":     req.Time,
		"type":     "TEXT",
	}

	return NotifMessageResponse(message)
}

func BuildMessageEditEvent(req MessageEditEventRequest) map[string]any {
	if req.Time == 0 {
		req.Time = time.Now().UnixMilli()
	}

	message := map[string]any{
		"id":       req.MessageID,
		"chatId":   req.ChatID,
		"text":     req.Text,
		"senderId": req.SenderID,
		"time":     req.Time,
		"type":     "TEXT",
		"status":   MessageStatusEdited,
	}

	return NotifMessageResponse(message)
}

func BuildMessageDeleteEvent(req MessageDeleteEventRequest) map[string]any {
	payload := map[string]any{
		"chatId": req.ChatID,
	}

	if req.MessageID > 0 {
		payload["messageId"] = req.MessageID
	}

	if len(req.MessageIDs) > 0 {
		payload["messageIds"] = req.MessageIDs
	}

	return map[string]any{
		"ver":     ProtocolVersion,
		"cmd":     ProtocolCommand,
		"seq":     0,
		"opcode":  OpcodeNotifMsgDelete,
		"payload": payload,
	}
}

func BuildReactionEvent(req ReactionEventRequest) map[string]any {
	if req.TotalCount == 0 && req.Reaction != "" {
		req.TotalCount = 1
	}

	if req.Counters == nil && req.Reaction != "" {
		req.Counters = []map[string]any{
			{
				"reactionType": ReactionTypeEmoji,
				"reaction":     req.Reaction,
				"count":        1,
			},
		}
	}

	return NotifReactionChangedResponse(req.ChatID, req.MessageID, req.TotalCount, req.YourReaction, req.Counters)
}

func BuildChatUpdateEvent(req ChatUpdateEventRequest) map[string]any {
	if req.Chat == nil {
		req.Chat = map[string]any{
			"id":   req.ChatID,
			"type": ChatTypeChat,
		}
	}

	if req.Data != nil {
		for k, v := range req.Data {
			req.Chat[k] = v
		}
	}

	return NotifChatResponse(req.Chat)
}

func BuildContactUpdateEvent(req ContactUpdateEventRequest) map[string]any {
	payload := map[string]any{}

	if req.Contact != nil {
		payload["contact"] = req.Contact
	}

	if req.Data != nil {
		for k, v := range req.Data {
			payload[k] = v
		}
	}

	return map[string]any{
		"ver":     ProtocolVersion,
		"cmd":     ProtocolCommand,
		"seq":     0,
		"opcode":  OpcodeNotifContact,
		"payload": payload,
	}
}

func BuildTypingEvent(req TypingEventRequest) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    0,
		"opcode": OpcodeNotifTyping,
		"payload": map[string]any{
			"chatId": req.ChatID,
			"userId": req.UserID,
			"typing": req.Typing,
		},
	}
}
