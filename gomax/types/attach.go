package types

import "github.com/fresh-milkshake/gomax/enums"

// Описывает фото‑вложение сообщения в терминах API Max.
// Структура повторяет модель PyMax в упрощённом виде и может быть расширена при необходимости.
type PhotoAttach struct {
	Type       enums.AttachType `json:"_type"`
	PhotoID    int64            `json:"photoId"`
	BaseURL    string           `json:"baseUrl"`
	Width      int              `json:"width"`
	Height     int              `json:"height"`
	PreviewRaw *string          `json:"previewData,omitempty"`
}

// Описывает видео‑вложение сообщения (включая размеры и длительность).
type VideoAttach struct {
	Type     enums.AttachType `json:"_type"`
	VideoID  int64            `json:"videoId"`
	Width    int              `json:"width"`
	Height   int              `json:"height"`
	Duration int              `json:"duration"`
}

// Описывает вложение произвольного файла.
type FileAttach struct {
	Type enums.AttachType `json:"_type"`
	ID   int64            `json:"id"`
	Name string           `json:"name"`
	Size int64            `json:"size"`
}

// Представляет служебное CONTROL‑вложение (например, события управления чатом).
type ControlAttach struct {
	Type  enums.AttachType `json:"_type"`
	Event string           `json:"event"`
	Extra map[string]any   `json:"-"`
}

// Описывает стикер‑вложение с метаданными и ссылками.
type StickerAttach struct {
	Type        enums.AttachType `json:"_type"`
	StickerID   int64            `json:"stickerId"`
	SetID       int64            `json:"setId"`
	URL         string           `json:"url"`
	LottieURL   string           `json:"lottieUrl"`
	Width       int              `json:"width"`
	Height      int              `json:"height"`
	Time        int64            `json:"time"`
	StickerType string           `json:"stickerType"`
	AuthorType  string           `json:"authorType"`
	Audio       bool             `json:"audio"`
	Tags        []string         `json:"tags,omitempty"`
}

// Описывает аудио‑ или голосовое вложение с метаданными и токеном.
type AudioAttach struct {
	Type                enums.AttachType `json:"_type"`
	AudioID             int64            `json:"audioId"`
	Duration            int              `json:"duration"`
	URL                 string           `json:"url"`
	Wave                string           `json:"wave"`
	TranscriptionStatus string           `json:"transcriptionStatus"`
	Token               string           `json:"token"`
}

// Обобщённый контейнер для всех типов вложений сообщения,
// позволяющий десериализовать произвольный attach из ответа API.
type Attach struct {
	Type       enums.AttachType `json:"_type"`
	Photo      *PhotoAttach     `json:"photo,omitempty"`
	Video      *VideoAttach     `json:"video,omitempty"`
	File       *FileAttach      `json:"file,omitempty"`
	Control    *ControlAttach   `json:"control,omitempty"`
	Sticker    *StickerAttach   `json:"sticker,omitempty"`
	Audio      *AudioAttach     `json:"audio,omitempty"`
	RawPayload any              `json:"-"`
}
