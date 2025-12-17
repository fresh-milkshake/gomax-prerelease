package types

import "github.com/fresh-milkshake/gomax/enums"

// Чат или группа.
type Chat struct {
	ID                    int64            `json:"id"`
	Type                  enums.ChatType   `json:"type"`
	Title                 *string          `json:"title,omitempty"`
	Owner                 int64            `json:"owner"`
	Access                enums.AccessType `json:"access"`
	ParticipantsCount     int              `json:"participantsCount"`
	Admins                []int64          `json:"admins,omitempty"`
	Description           *string          `json:"description,omitempty"`
	Link                  *string          `json:"link,omitempty"`
	BaseIconURL           *string          `json:"baseIconUrl,omitempty"`
	BaseRawIconURL        *string          `json:"baseRawIconUrl,omitempty"`
	CID                   int64            `json:"cid"`
	Created               int64            `json:"created"`
	JoinTime              int64            `json:"joinTime"`
	LastMessage           *Message         `json:"lastMessage,omitempty"`
	LastEventTime         int64            `json:"lastEventTime"`
	LastDelayedUpdateTime int64            `json:"lastDelayedUpdateTime"`
	MessagesCount         int              `json:"messagesCount"`
	Modified              int64            `json:"modified"`
	Options               map[string]bool  `json:"options,omitempty"`
	PrevMessageID         *string          `json:"prevMessageId,omitempty"`
	Restrictions          *int             `json:"restrictions,omitempty"`
	Status                string           `json:"status"`
}

// Личный диалог.
type Dialog struct {
	ID                    int64          `json:"id"`
	Owner                 int64          `json:"owner"`
	Type                  enums.ChatType `json:"type"`
	LastMessage           *Message       `json:"lastMessage,omitempty"`
	CID                   *int64         `json:"cid,omitempty"`
	HasBots               *bool          `json:"hasBots,omitempty"`
	JoinTime              int64          `json:"joinTime"`
	Created               int64          `json:"created"`
	LastFireDelayedError  int64          `json:"lastFireDelayedErrorTime"`
	LastDelayedUpdateTime int64          `json:"lastDelayedUpdateTime"`
	PrevMessageID         *string        `json:"prevMessageId,omitempty"`
	Options               map[string]any `json:"options,omitempty"`
	Modified              int64          `json:"modified"`
	LastEventTime         int64          `json:"lastEventTime"`
	Status                string         `json:"status"`
}

// Специализированный тип чата.
type Channel = Chat
