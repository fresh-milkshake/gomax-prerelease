package payloads

import (
	"strings"
)

// Описывает базовое сообщение WebSocket‑протокола Max,
// содержащее версию, команду, порядковый номер и arbitrary payload.
type BaseWebSocketMessage struct {
	Ver     int            `json:"ver"`
	Cmd     int            `json:"cmd"`
	Seq     int            `json:"seq"`
	Opcode  int            `json:"opcode"`
	Payload map[string]any `json:"payload"`
}

// Преобразует snake_case идентификатор в camelCase
// и используется для приведения имён полей к стилю JSON‑ключей Max.
func toCamel(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return s
	}
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}
