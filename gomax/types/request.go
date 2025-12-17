package types

import "encoding/json"

// Запрос файла (метаданные).
type FileRequest struct {
	Unsafe bool   `json:"unsafe"`
	URL    string `json:"url"`
}

// Запрос видео (метаданные).
// В Python версии URL определяется динамически из полей, не равных "EXTERNAL" и "cache".
type VideoRequest struct {
	External string `json:"EXTERNAL"`
	Cache    bool   `json:"cache"`
	URL      string `json:"-"`
}

// Реализует кастомную десериализацию для VideoRequest.
func (v *VideoRequest) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if ext, ok := m["EXTERNAL"].(string); ok {
		v.External = ext
	}
	if cache, ok := m["cache"].(bool); ok {
		v.Cache = cache
	}

	for k, val := range m {
		if k != "EXTERNAL" && k != "cache" {
			if url, ok := val.(string); ok {
				v.URL = url
				break
			}
		}
	}

	return nil
}
