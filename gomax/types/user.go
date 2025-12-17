package types

// Описывает текущего пользователя.
type Me struct {
	ID            int64    `json:"id"`
	Phone         string   `json:"phone"`
	Names         []Names  `json:"names,omitempty"`
	AccountStatus int      `json:"accountStatus"`
	UpdateTime    int64    `json:"updateTime"`
	Options       []string `json:"options,omitempty"`
}

// Описывает произвольного пользователя.
type User struct {
	ID            int64    `json:"id"`
	Names         []Names  `json:"names,omitempty"`
	AccountStatus int      `json:"accountStatus"`
	UpdateTime    int64    `json:"updateTime"`
	Options       []string `json:"options,omitempty"`
	BaseURL       *string  `json:"baseUrl,omitempty"`
	BaseRawURL    *string  `json:"baseRawUrl,omitempty"`
	PhotoID       *int64   `json:"photoId,omitempty"`
	Description   *string  `json:"description,omitempty"`
	Gender        *int     `json:"gender,omitempty"`
	Link          *string  `json:"link,omitempty"`
	WebApp        *string  `json:"webApp,omitempty"`
}

// Запись в адресной книге.
type Contact struct {
	ID            int64    `json:"id"`
	Names         []Names  `json:"names,omitempty"`
	AccountStatus int      `json:"accountStatus"`
	PhotoID       *int64   `json:"photoId,omitempty"`
	BaseURL       *string  `json:"baseUrl,omitempty"`
	BaseRawURL    *string  `json:"baseRawUrl,omitempty"`
	Options       []string `json:"options,omitempty"`
	UpdateTime    int64    `json:"updateTime"`
}

// Участник чата.
type Member struct {
	Contact  Contact   `json:"contact"`
	Presence *Presence `json:"presence,omitempty"`
	ReadMark int64     `json:"readMark"`
}
