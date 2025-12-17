package mockserver

import "time"

func buildMessageEventFromMap(req map[string]any) map[string]any {
	chatID, _ := req["chatId"].(float64)
	messageID, _ := req["messageId"].(float64)
	text, _ := req["text"].(string)
	senderID, _ := req["senderId"].(float64)
	timeVal, ok := req["time"].(float64)
	if !ok {
		timeVal = float64(time.Now().UnixMilli())
	}

	message := map[string]any{
		"id":       int64(messageID),
		"chatId":   int64(chatID),
		"text":     text,
		"senderId": int64(senderID),
		"time":     int64(timeVal),
	}

	return NotifMessageResponse(message)
}

func buildMessageEditEventFromMap(req map[string]any) map[string]any {
	chatID, _ := req["chatId"].(float64)
	messageID, _ := req["messageId"].(float64)
	text, _ := req["text"].(string)
	senderID, _ := req["senderId"].(float64)
	timeVal, ok := req["time"].(float64)
	if !ok {
		timeVal = float64(time.Now().UnixMilli())
	}

	message := map[string]any{
		"id":       int64(messageID),
		"chatId":   int64(chatID),
		"text":     text,
		"senderId": int64(senderID),
		"time":     int64(timeVal),
		"status":   MessageStatusEdited,
	}

	return NotifMessageResponse(message)
}

func buildMessageDeleteEventFromMap(req map[string]any) map[string]any {
	chatID, _ := req["chatId"].(float64)
	payload := map[string]any{
		"chatId": int64(chatID),
	}

	if messageID, ok := req["messageId"].(float64); ok && messageID > 0 {
		payload["messageId"] = int64(messageID)
	}

	if messageIDs, ok := req["messageIds"].([]any); ok && len(messageIDs) > 0 {
		ids := make([]int64, 0, len(messageIDs))
		for _, id := range messageIDs {
			if idFloat, ok := id.(float64); ok {
				ids = append(ids, int64(idFloat))
			}
		}
		if len(ids) > 0 {
			payload["messageIds"] = ids
		}
	}

	return map[string]any{
		"ver":     ProtocolVersion,
		"cmd":     ProtocolCommand,
		"seq":     0,
		"opcode":  OpcodeNotifMsgDelete,
		"payload": payload,
	}
}

func buildReactionEventFromMap(req map[string]any) map[string]any {
	chatID, _ := req["chatId"].(float64)
	messageID, _ := req["messageId"].(string)
	reaction, _ := req["reaction"].(string)
	totalCount, _ := req["totalCount"].(float64)
	yourReaction, _ := req["yourReaction"].(string)

	var counters []map[string]any
	if countersRaw, ok := req["counters"].([]any); ok && len(countersRaw) > 0 {
		counters = make([]map[string]any, 0, len(countersRaw))
		for _, c := range countersRaw {
			if counterMap, ok := c.(map[string]any); ok {
				counters = append(counters, counterMap)
			}
		}
	}

	if totalCount == 0 && reaction != "" {
		totalCount = 1
	}

	if counters == nil && reaction != "" {
		counters = []map[string]any{
			{
				"reactionType": ReactionTypeEmoji,
				"id":           reaction,
				"count":        1,
			},
		}
	}

	return NotifReactionChangedResponse(int64(chatID), messageID, int(totalCount), yourReaction, counters)
}

func buildChatUpdateEventFromMap(req map[string]any) map[string]any {
	chatID, _ := req["chatId"].(float64)
	chat, _ := req["chat"].(map[string]any)
	data, _ := req["data"].(map[string]any)

	if chat == nil {
		chat = map[string]any{
			"id":   int64(chatID),
			"type": ChatTypeChat,
		}
	}

	for k, v := range data {
		chat[k] = v
	}

	return NotifChatResponse(chat)
}

func buildContactUpdateEventFromMap(req map[string]any) map[string]any {
	contact, _ := req["contact"].(map[string]any)
	data, _ := req["data"].(map[string]any)

	payload := map[string]any{}

	if contact != nil {
		payload["contact"] = contact
	}

	if len(data) > 0 {
		for k, v := range data {
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

func buildTypingEventFromMap(req map[string]any) map[string]any {
	chatID, _ := req["chatId"].(float64)
	userID, _ := req["userId"].(float64)
	typing, _ := req["typing"].(bool)

	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    0,
		"opcode": OpcodeNotifTyping,
		"payload": map[string]any{
			"chatId": int64(chatID),
			"userId": int64(userID),
			"typing": typing,
		},
	}
}
