package payloads

import "github.com/fresh-milkshake/gomax/enums"

// Описывает ссылку на сообщение, на которое оформляется ответ.
type ReplyLink struct {
	Type      string `json:"type"`
	MessageID string `json:"messageId"`
}

// Описывает запрос на резервирование слота загрузки файлов/медиа.
type UploadPayload struct {
	Count int `json:"count"`
}

// Описывает вложение фото, отправляемое вместе с сообщением.
type AttachPhotoPayload struct {
	Type       enums.AttachType `json:"_type"`
	PhotoToken string           `json:"photoToken"`
}

// Описывает вложение видео, отправляемое вместе с сообщением.
type VideoAttachPayload struct {
	Type    enums.AttachType `json:"_type"`
	VideoID int64            `json:"videoId"`
	Token   string           `json:"token"`
}

// Описывает вложение произвольного файла в сообщении.
type AttachFilePayload struct {
	Type   enums.AttachType `json:"_type"`
	FileID int64            `json:"fileId"`
}

// Описывает элемент форматирования текста (жирный, курсив и т.п.) в payload сообщения.
type MessageElement struct {
	Type   string `json:"type"`
	From   int    `json:"from"`
	Length int    `json:"length"`
}

// Описывает внутреннюю структуру отправляемого сообщения
// с текстом, форматированием, вложениями и ссылкой‑ответом.
type SendMessagePayloadMessage struct {
	Text     string           `json:"text"`
	CID      int64            `json:"cid"`
	Elements []MessageElement `json:"elements"`
	Attaches []interface{}    `json:"attaches"`
	Link     *ReplyLink       `json:"link,omitempty"`
}

// Описывает payload WebSocket‑команды отправки сообщения.
type SendMessagePayload struct {
	ChatID  int64                     `json:"chatId"`
	Message SendMessagePayloadMessage `json:"message"`
	Notify  bool                      `json:"notify"`
}

// Описывает payload команды редактирования существующего сообщения.
type EditMessagePayload struct {
	ChatID    int64            `json:"chatId"`
	MessageID int64            `json:"messageId"`
	Text      string           `json:"text"`
	Elements  []MessageElement `json:"elements"`
	Attaches  []interface{}    `json:"attaches"`
}

// Описывает payload команды удаления одного или нескольких сообщений.
type DeleteMessagePayload struct {
	ChatID     int64   `json:"chatId"`
	MessageIDs []int64 `json:"messageIds"`
	ForMe      bool    `json:"forMe"`
}

// Описывает запрос истории сообщений чата
// с параметрами окна (forward/backward) и флагом получения самих сообщений.
type FetchHistoryPayload struct {
	ChatID      int64  `json:"chatId"`
	FromTime    *int64 `json:"from,omitempty"`
	Forward     int    `json:"forward"`
	Backward    int    `json:"backward"`
	GetMessages bool   `json:"getMessages"`
}

// Описывает запрос на закрепление сообщения в чате.
type PinMessagePayload struct {
	ChatID       int64 `json:"chatId"`
	NotifyPin    bool  `json:"notifyPin"`
	PinMessageID int64 `json:"pinMessageId"`
}

// Описывает запрос на получение метаданных видео по идентификаторам.
type GetVideoPayload struct {
	ChatID    int64       `json:"chatId"`
	MessageID interface{} `json:"messageId"`
	VideoID   int64       `json:"videoId"`
}

// Описывает запрос на получение метаданных файла по идентификаторам.
type GetFilePayload struct {
	ChatID    int64       `json:"chatId"`
	MessageID interface{} `json:"messageId"`
	FileID    int64       `json:"fileId"`
}

// Описывает одну реакцию‑эмодзи, которую необходимо применить к сообщению.
type ReactionInfoPayload struct {
	ReactionType string `json:"reactionType"`
	ID           string `json:"id"`
}

// Описывает команду добавления реакции к сообщению.
type AddReactionPayload struct {
	ChatID    int64               `json:"chatId"`
	MessageID string              `json:"messageId"`
	Reaction  ReactionInfoPayload `json:"reaction"`
}

// Описывает запрос на получение агрегированных реакций для сообщений.
type GetReactionsPayload struct {
	ChatID     int64    `json:"chatId"`
	MessageIDs []string `json:"messageIds"`
}

// Описывает команду удаления реакции текущего пользователя.
type RemoveReactionPayload struct {
	ChatID    int64  `json:"chatId"`
	MessageID string `json:"messageId"`
}
