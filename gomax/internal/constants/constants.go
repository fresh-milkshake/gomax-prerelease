package constants

import "regexp"

// Набор значений по умолчанию и констант протокола Max, согласованный с Python‑версией PyMax.
var (
	PhoneRegex                   = regexp.MustCompile(`^\+?\d{10,15}$`)
	WebsocketURI                 = "wss://ws-api.oneme.ru/websocket"
	WebsocketOrigin              = "https://web.max.ru"
	Host                         = "api.oneme.ru"
	Port                     int = 443
	DefaultTimeout               = 10.0
	DefaultDeviceType            = "WEB"
	DefaultLocale                = "ru"
	DefaultDeviceLocale          = "ru"
	DefaultDeviceName            = "Chrome"
	DefaultAppVersion            = "25.10.13"
	DefaultScreen                = "1080x1920 1.0x"
	DefaultOSVersion             = "Linux"
	DefaultUserAgent             = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36"
	DefaultTimezone              = "Europe/Moscow"
	DefaultChatMembers           = 50
	DefaultMarker                = 0
	DefaultPingInterval          = 30.0
	RecvLoopBackoff              = 0.5
	DefaultMaxRetries            = 3
	DefaultRetryInitialDelay     = 1.0
	DefaultRetryMaxDelay         = 30.0

	// ProtocolVersion версия протокола WebSocket API Max
	ProtocolVersion = 11

	// ProtocolCommand команда протокола (0 = request)
	ProtocolCommand = 0

	// StatusOK статус успешного выполнения
	StatusOK = "ok"

	// ChatTypeChat тип группового чата
	ChatTypeChat = "CHAT"

	// ReactionTypeEmoji тип реакции - эмодзи
	ReactionTypeEmoji = "EMOJI"

	// AttachTypeControl тип вложения - управляющее сообщение
	AttachTypeControl = "CONTROL"

	// LinkTypeReply тип ссылки - ответ на сообщение
	LinkTypeReply = "REPLY"

	// MemberTypeMember тип участника
	MemberTypeMember = "MEMBER"

	// OperationAdd операция добавления
	OperationAdd = "add"

	// OperationRemove операция удаления
	OperationRemove = "remove"

	// TokenTypeRegister тип токена - регистрация
	TokenTypeRegister = "REGISTER"

	// TokenTypeLogin тип токена - вход
	TokenTypeLogin = "LOGIN"

	// NameTypeFirstLast тип имени - имя и фамилия
	NameTypeFirstLast = "FIRST_LAST"

	// EventNew событие создания
	EventNew = "new"

	// MessageStatusEdited статус сообщения - отредактировано
	MessageStatusEdited = "EDITED"
)
