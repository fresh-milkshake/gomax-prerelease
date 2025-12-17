package enums

// Описывает этапы и типы токенов аутентификации в протоколе Max.
// Значения соответствуют одноимённым константам в PyMax.
type AuthType string

const (
	AuthTypeStartAuth AuthType = "START_AUTH"
	AuthTypeCheckCode AuthType = "CHECK_CODE"
	AuthTypeRegister  AuthType = "REGISTER"
)
