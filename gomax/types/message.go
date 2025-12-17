package types

import (
	"encoding/json"
	"strconv"

	"github.com/fresh-milkshake/gomax/enums"
)

// Связь на сообщение в другом чате/диалоге.
type MessageLink struct {
	ChatID  int64   `json:"chatId"`
	Type    string  `json:"type"`
	Message Message `json:"message"`
}

// Счётчик конкретной реакции.
type ReactionCounter struct {
	Count    int    `json:"count"`
	Reaction string `json:"reaction"`
}

// Информация о реакциях сообщения.
type ReactionInfo struct {
	TotalCount   int               `json:"totalCount"`
	YourReaction *string           `json:"yourReaction,omitempty"`
	Counters     []ReactionCounter `json:"counters,omitempty"`
}

// Сообщение в чате.
type Message struct {
	ID        int64                `json:"id"`
	ChatID    *int64               `json:"chatId,omitempty"`
	Sender    *int64               `json:"sender,omitempty"`
	Text      string               `json:"text"`
	Time      int64                `json:"time"`
	Type      enums.MessageType    `json:"type"`
	Elements  []Element            `json:"elements,omitempty"`
	Attaches  []Attach             `json:"attaches,omitempty"`
	Status    *enums.MessageStatus `json:"status,omitempty"`
	Reactions *ReactionInfo        `json:"reactionInfo,omitempty"`
	Link      *MessageLink         `json:"link,omitempty"`
	Options   *int                 `json:"options,omitempty"`
}

// UnmarshalJSON кастомно парсит Message, обрабатывая ID как строку или число.
func (m *Message) UnmarshalJSON(data []byte) error {

	type Alias Message
	aux := &struct {
		ID interface{} `json:"id"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.ID.(type) {
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		m.ID = id
	case float64:
		m.ID = int64(v)
	case int64:
		m.ID = v
	case int:
		m.ID = int64(v)
	default:

		m.ID = 0
	}

	return nil
}

// Описывает фрагмент форматированного текста.
// Используется в payload-ах для отправки сообщений.
type MessageElement struct {
	Type   string `json:"type"`
	From   int    `json:"from"`
	Length int    `json:"length"`
}
