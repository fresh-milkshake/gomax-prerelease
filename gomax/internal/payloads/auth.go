package payloads

import "github.com/fresh-milkshake/gomax/enums"

// Payload для запроса кода аутентификации.
type RequestCodePayload struct {
	Phone    string         `json:"phone"`
	Type     enums.AuthType `json:"type"`
	Language string         `json:"language"`
}

// Payload для отправки кода верификации.
type SendCodePayload struct {
	Token         string         `json:"token"`
	VerifyCode    string         `json:"verifyCode"`
	AuthTokenType enums.AuthType `json:"authTokenType"`
}

// Payload для регистрации нового пользователя.
type RegisterPayload struct {
	FirstName string         `json:"firstName"`
	LastName  *string        `json:"lastName,omitempty"`
	Token     string         `json:"token"`
	TokenType enums.AuthType `json:"tokenType"`
}

// Payload для синхронизации состояния.
type SyncPayload struct {
	Interactive  bool   `json:"interactive"`
	Token        string `json:"token"`
	ChatsSync    int    `json:"chatsSync"`
	ContactsSync int    `json:"contactsSync"`
	PresenceSync int    `json:"presenceSync"`
	DraftsSync   int    `json:"draftsSync"`
	ChatsCount   int    `json:"chatsCount"`
}
