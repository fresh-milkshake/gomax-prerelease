package filters

import (
	"github.com/fresh-milkshake/gomax/enums"
	"github.com/fresh-milkshake/gomax/types"
)

// Описывает критерии фильтрации сообщений для обработчиков событий MaxClient.
type Filter struct {
	ChatID       *int64
	UserID       *int64
	Text         []string
	Status       *enums.MessageStatus
	Type         *enums.MessageType
	TextContains *string
	ReactionInfo *bool
}

// Проверяет, соответствует ли сообщение всем критериям фильтра,
// включая чат, отправителя, текст, статус, тип и наличие реакций.
func (f *Filter) Match(msg *types.Message) bool {
	if f.ChatID != nil && (msg.ChatID == nil || *msg.ChatID != *f.ChatID) {
		return false
	}
	if f.UserID != nil && (msg.Sender == nil || *msg.Sender != *f.UserID) {
		return false
	}
	if f.Text != nil {
		for _, t := range f.Text {
			if !contains(msg.Text, t) {
				return false
			}
		}
	}
	if f.TextContains != nil && !contains(msg.Text, *f.TextContains) {
		return false
	}
	if f.Status != nil && (msg.Status == nil || *msg.Status != *f.Status) {
		return false
	}
	if f.Type != nil && msg.Type != *f.Type {
		return false
	}
	if f.ReactionInfo != nil {
		hasReactions := msg.Reactions != nil
		if *f.ReactionInfo && !hasReactions {
			return false
		}
		if !*f.ReactionInfo && hasReactions {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
