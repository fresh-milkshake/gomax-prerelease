package types

import "github.com/fresh-milkshake/gomax/enums"

// Описывает элемент форматированного текста.
type Element struct {
	Type   enums.FormattingType `json:"type"`
	From   *int                 `json:"from,omitempty"`
	Length int                  `json:"length"`
}
