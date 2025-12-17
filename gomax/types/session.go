package types

// Информация о сессии пользователя.
type Session struct {
	Client   string `json:"client"`
	Info     string `json:"info"`
	Location string `json:"location"`
	Time     int64  `json:"time"`
	Current  *bool  `json:"current,omitempty"`
}
