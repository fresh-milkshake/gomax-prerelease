package utils

import (
	"encoding/json"
	"reflect"
)

// Преобразует произвольную структуру в map[string]any
// с учётом JSON‑тегов для последующей отправки в WebSocket API.
func ToMap(v interface{}) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Заполняет структуру из map[string]any, используя JSON‑кодек для маппинга полей.
func FromMap(m map[string]any, v interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Преобразует слайс структур в слайс map[string]any
// и удобен для сериализации коллекций payload‑ов.
func ToMapSlice(v interface{}) ([]map[string]any, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil, nil
	}

	result := make([]map[string]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		m, err := ToMap(item)
		if err != nil {
			return nil, err
		}
		result[i] = m
	}
	return result, nil
}
