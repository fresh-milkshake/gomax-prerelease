package enums

// Описывает тип чата: личный диалог, групповая беседа или канал.
type ChatType string

const (
	ChatTypeDialog  ChatType = "DIALOG"
	ChatTypeChat    ChatType = "CHAT"
	ChatTypeChannel ChatType = "CHANNEL"
)

// Описывает тип сообщения (обычное, системное или сервисное).
type MessageType string

const (
	MessageTypeText    MessageType = "TEXT"
	MessageTypeSystem  MessageType = "SYSTEM"
	MessageTypeService MessageType = "SERVICE"
)

// Описывает статус сообщения: отредактировано или удалено.
type MessageStatus string

const (
	MessageStatusEdited  MessageStatus = "EDITED"
	MessageStatusRemoved MessageStatus = "REMOVED"
)

// Описывает уровень доступа к чату или каналу.
type AccessType string

const (
	AccessTypePublic  AccessType = "PUBLIC"
	AccessTypePrivate AccessType = "PRIVATE"
	AccessTypeSecret  AccessType = "SECRET"
)

// Описывает тип вложения сообщения: фото, видео, файл, стикер, аудио или управляющее.
type AttachType string

const (
	AttachTypePhoto   AttachType = "PHOTO"
	AttachTypeVideo   AttachType = "VIDEO"
	AttachTypeFile    AttachType = "FILE"
	AttachTypeSticker AttachType = "STICKER"
	AttachTypeAudio   AttachType = "AUDIO"
	AttachTypeControl AttachType = "CONTROL"
)
