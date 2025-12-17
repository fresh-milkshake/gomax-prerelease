package gomax

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

// Описывает ошибку протокола Max с деталями,
// возвращёнными сервером WebSocket API.
type Error struct {
	Code             string
	Message          string
	Title            string
	LocalizedMessage string
}

// Форматирует ошибку Max в читабельную строку для интерфейса error.
func (e *Error) Error() string {
	msg := "gomax error"
	if e.LocalizedMessage != "" {
		msg += ": " + e.LocalizedMessage
	}
	if e.Message != "" {
		msg += ": " + e.Message
	}
	if e.Title != "" {
		msg += fmt.Sprintf(" (%s)", e.Title)
	}
	if e.Code != "" {
		msg += fmt.Sprintf(" [%s]", e.Code)
	}
	return msg
}

// Сигнализирует о превышении лимита запросов к API Max (HTTP 429/too.many.requests).
type RateLimitError struct {
	Err *Error
}

// Возвращает текст вложенной ошибки Max либо общий текст про лимиты.
func (e *RateLimitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "rate limit error"
}

// Описывает неудачную авторизацию пользователя в Max.
type LoginError struct {
	*Error
}

// Сигнализирует о логической ошибке в ответе сервера,
// когда структура ответа корректна, но семантика не соответствует ожиданиям клиента.
type ResponseError struct {
	Message string
}

// Возвращает человекочитаемое описание логической ошибки.
func (e *ResponseError) Error() string {
	return "response error: " + e.Message
}

// Сигнализирует о неожиданной структуре ответа WebSocket API Max.
type ResponseStructureError struct {
	Message string
}

// Возвращает текстовое описание проблемы со структурой ответа.
func (e *ResponseStructureError) Error() string {
	return "response structure error: " + e.Message
}

// Указывает на неверный формат номера телефона при создании клиента.
type InvalidPhoneError struct {
	Phone string
}

// Возвращает сообщение об ошибочном формате номера.
func (e *InvalidPhoneError) Error() string {
	return fmt.Sprintf("invalid phone number format: %s", e.Phone)
}

// Возвращается при попытке отправки или чтения,
// когда WebSocket‑соединение ещё не установлено или уже закрыто.
type WebSocketNotConnectedError struct{}

// Возвращает текстовое описание отсутствия активного WebSocket‑соединения.
func (WebSocketNotConnectedError) Error() string {
	return "websocket is not connected"
}

// Возвращается при работе с неинициализированным низкоуровневым сокетом.
type SocketNotConnectedError struct{}

// Возвращает текстовое описание отсутствия соединения на уровне сокета.
func (SocketNotConnectedError) Error() string {
	return "socket is not connected"
}

// Возвращается при ошибке отправки данных по сырому сокету и ожидании ответа.
type SocketSendError struct{}

// Возвращает текстовое описание ошибки отправки по сокету.
func (SocketSendError) Error() string {
	return "send and wait failed (socket)"
}

// Сигнализирует о сетевой ошибке при выполнении HTTP или WebSocket операций.
type NetworkError struct {
	Err error
}

// Возвращает текстовое описание сетевой ошибки.
func (e *NetworkError) Error() string {
	if e.Err != nil {
		return "network error: " + e.Err.Error()
	}
	return "network error"
}

// Unwrap возвращает вложенную ошибку для поддержки errors.Is и errors.As.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// Сигнализирует о временной сетевой ошибке, которую можно повторить.
type TemporaryError struct {
	Err error
}

// Возвращает текстовое описание временной ошибки.
func (e *TemporaryError) Error() string {
	if e.Err != nil {
		return "temporary error: " + e.Err.Error()
	}
	return "temporary error"
}

// Unwrap возвращает вложенную ошибку для поддержки errors.Is и errors.As.
func (e *TemporaryError) Unwrap() error {
	return e.Err
}

// Temporary возвращает true, так как это временная ошибка.
func (e *TemporaryError) Temporary() bool {
	return true
}

// IsTemporaryError проверяет, является ли ошибка временной и может быть повторена.
func IsTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	var tempErr *TemporaryError
	if errors.As(err, &tempErr) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Temporary() || netErr.Timeout()
	}

	var sysErr syscall.Errno
	if errors.As(err, &sysErr) {
		return sysErr == syscall.ECONNRESET || sysErr == syscall.ECONNREFUSED ||
			sysErr == syscall.ETIMEDOUT || sysErr == syscall.EAGAIN
	}

	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		if IsTemporaryError(unwrapped) {
			return true
		}
	}

	errStr := err.Error()
	return containsAny(errStr, "timeout", "connection reset", "connection refused",
		"temporary failure", "network is unreachable", "no route to host")
}

// IsNetworkError проверяет, является ли ошибка сетевой.
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return true
	}

	var netErrInterface net.Error
	if errors.As(err, &netErrInterface) {
		return true
	}

	errStr := err.Error()
	return containsAny(errStr, "network", "connection", "socket", "dial", "read", "write")
}

func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// Разбирает стандартный error‑payload WebSocket API Max
// и формирует подходящую Go‑ошибку (включая RateLimitError для too.many.requests).
// Возвращает nil, если ошибки нет.
func HandleError(data map[string]any) error {
	payload, ok := data["payload"].(map[string]any)
	if !ok {
		return &ResponseError{Message: "invalid response structure"}
	}

	errCode, ok := payload["error"].(string)
	if !ok || errCode == "" {

		return nil
	}

	message, _ := payload["message"].(string)
	title, _ := payload["title"].(string)
	localizedMessage, _ := payload["localizedMessage"].(string)

	if errCode == "too.many.requests" {
		return &RateLimitError{
			Err: &Error{
				Code:             errCode,
				Message:          message,
				Title:            title,
				LocalizedMessage: localizedMessage,
			},
		}
	}

	return &Error{
		Code:             errCode,
		Message:          message,
		Title:            title,
		LocalizedMessage: localizedMessage,
	}
}
