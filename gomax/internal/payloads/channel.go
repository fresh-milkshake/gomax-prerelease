package payloads

// Payload для разрешения ссылки.
type ResolveLinkPayload struct {
	Link string `json:"link"`
}

// Payload для присоединения к чату/каналу.
type JoinChatPayload struct {
	Link string `json:"link"`
}

// Payload для получения участников группы.
type GetGroupMembersPayload struct {
	Type   string `json:"type"`
	Marker *int   `json:"marker,omitempty"`
	ChatID int64  `json:"chatId"`
	Count  int    `json:"count"`
}

// Payload для поиска участников группы.
type SearchGroupMembersPayload struct {
	Type   string `json:"type"`
	Query  string `json:"query"`
	ChatID int64  `json:"chatId"`
}
