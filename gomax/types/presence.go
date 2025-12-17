package types

// Описывает информацию о присутствии пользователя (последнее время онлайн).
// Структура соответствует Presence из PyMax.
type Presence struct {
	Seen *int64 `json:"seen,omitempty"`
}

// Описывает одно представление имени пользователя, возвращаемое API Max.
type Name struct {
	Name      *string `json:"name,omitempty"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
	Type      *string `json:"type,omitempty"`
}

// Синоним для Name, соответствующий одноимённому типу в Python‑версии.
type Names = Name
