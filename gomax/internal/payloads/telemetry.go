package payloads

// Параметры события навигации.
type NavigationEventParams struct {
	ActionID   int  `json:"actionId"`
	ScreenTo   int  `json:"screenTo"`
	ScreenFrom *int `json:"screenFrom,omitempty"`
	SourceID   int  `json:"sourceId"`
	SessionID  int  `json:"sessionId"`
}

// Payload события навигации.
type NavigationEventPayload struct {
	Event  string                `json:"event"`
	Time   int64                 `json:"time"`
	Type   string                `json:"type"`
	UserID int64                 `json:"userId"`
	Params NavigationEventParams `json:"params"`
}

// Payload для отправки событий навигации.
type NavigationPayload struct {
	Events []NavigationEventPayload `json:"events"`
}

// Payload для User-Agent заголовков.
type UserAgentPayload struct {
	DeviceType      string `json:"deviceType"`
	Locale          string `json:"locale"`
	DeviceLocale    string `json:"deviceLocale"`
	OSVersion       string `json:"osVersion"`
	DeviceName      string `json:"deviceName"`
	HeaderUserAgent string `json:"headerUserAgent"`
	AppVersion      string `json:"appVersion"`
	Screen          string `json:"screen"`
	Timezone        string `json:"timezone"`
	ClientSessionID int    `json:"clientSessionId"`
	BuildNumber     int    `json:"buildNumber"`
}
