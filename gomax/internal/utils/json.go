package utils

import "encoding/json"

// Обёртка над json.Marshal, используемая для единообразного управления кодированием.
func JSONMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Обёртка над json.Unmarshal для декодирования произвольных структур из JSON.
func JSONUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
