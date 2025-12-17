package enums

// Тип элемента форматированного текста.
type ElementType string

const (
	ElementTypeText    ElementType = "text"
	ElementTypeMention ElementType = "mention"
	ElementTypeLink    ElementType = "link"
	ElementTypeEmoji   ElementType = "emoji"
)

// Тип форматирования текста.
type FormattingType string

const (
	FormattingTypeStrong        FormattingType = "STRONG"
	FormattingTypeEmphasized    FormattingType = "EMPHASIZED"
	FormattingTypeUnderline     FormattingType = "UNDERLINE"
	FormattingTypeStrikethrough FormattingType = "STRIKETHROUGH"
)

// Тип разметки сообщения (символы для markdown).
type MarkupType string

const (
	MarkupTypeBold          MarkupType = "**"
	MarkupTypeItalic        MarkupType = "*"
	MarkupTypeUnderline     MarkupType = "__"
	MarkupTypeStrikethrough MarkupType = "~~"
)

// Тип устройства платформы.
type DeviceType string

const (
	DeviceTypeWeb     DeviceType = "WEB"
	DeviceTypeAndroid DeviceType = "ANDROID"
	DeviceTypeIOS     DeviceType = "IOS"
)

// Действие взаимодействия с контактом.
type ContactAction string

const (
	ContactActionAdd    ContactAction = "ADD"
	ContactActionRemove ContactAction = "REMOVE"
)
