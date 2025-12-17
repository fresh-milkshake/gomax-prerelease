package payloads

import "github.com/fresh-milkshake/gomax/enums"

// Payload для получения информации о контактах.
type FetchContactsPayload struct {
	ContactIDs []int64 `json:"contactIds"`
}

// Payload для поиска пользователя по телефону.
type SearchByPhonePayload struct {
	Phone string `json:"phone"`
}

// Payload для действий с контактом.
type ContactActionPayload struct {
	ContactID int64               `json:"contactId"`
	Action    enums.ContactAction `json:"action"`
}
