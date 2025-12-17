package mockserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestContext создаёт контекст с таймаутом для тестов.
func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// TestUser создаёт тестового пользователя.
func TestUser(id int64, firstName, lastName, phone string) map[string]any {
	return map[string]any{
		"id":    id,
		"phone": phone,
		"names": []map[string]any{
			{
				"firstName": firstName,
				"lastName":  lastName,
				"type":      "FIRST_LAST",
			},
		},
	}
}

// TestChat создаёт тестовый чат.
func TestChat(id int64, chatType, title string) map[string]any {
	return map[string]any{
		"id":    id,
		"type":  chatType,
		"title": title,
	}
}

// TestMessage создаёт тестовое сообщение.
func TestMessage(id, chatID, senderID int64, text string) map[string]any {
	return map[string]any{
		"id":       id,
		"chatId":   chatID,
		"senderId": senderID,
		"text":     text,
		"time":     time.Now().UnixMilli(),
	}
}

// TestSession создаёт тестовую сессию.
func TestSession(id string, deviceType, platform string, current bool) map[string]any {
	return map[string]any{
		"id":         id,
		"deviceType": deviceType,
		"platform":   platform,
		"current":    current,
		"lastActive": time.Now().UnixMilli(),
	}
}

// TestFolder создаёт тестовую папку.
func TestFolder(id, title string, include []int64) map[string]any {
	return map[string]any{
		"id":      id,
		"title":   title,
		"include": include,
	}
}

// TestMember создаёт тестового участника группы/канала.
func TestMember(userID int64, role string) map[string]any {
	return map[string]any{
		"userId": userID,
		"role":   role,
	}
}

// TestContact создаёт тестовый контакт.
func TestContact(id int64, firstName, lastName, phone string) map[string]any {
	return map[string]any{
		"id":    id,
		"phone": phone,
		"names": []map[string]any{
			{
				"firstName": firstName,
				"lastName":  lastName,
				"type":      "FIRST_LAST",
			},
		},
	}
}

// TestReactionCounter создаёт тестовый счётчик реакций.
func TestReactionCounter(reactionType, id string, count int) map[string]any {
	return map[string]any{
		"reactionType": reactionType,
		"id":           id,
		"count":        count,
	}
}

// StartMockServer создаёт и запускает mock сервер для теста.
func StartMockServer(t *testing.T) *MockServer {
	server, err := NewMockServer()
	require.NoError(t, err, "failed to create mock server")

	server.Start()

	t.Cleanup(func() {
		server.Stop()
	})

	return server
}

// StartMockServerWithDefaults создаёт mock сервер со стандартными обработчиками.
func StartMockServerWithDefaults(t *testing.T) *MockServer {
	server := StartMockServer(t)
	server.DefaultHandlers()
	return server
}

// MockHTTPServer создаёт mock HTTP сервер для тестирования загрузки файлов.
func MockHTTPServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

// PhotoUploadHandler создаёт обработчик для тестирования загрузки фото.
func PhotoUploadHandler(photoToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"photos":{"1":{"token":"` + photoToken + `"}}}`))
	}
}

// FileUploadHandler создаёт обработчик для тестирования загрузки файлов.
func FileUploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// WaitForCondition ждёт выполнения условия с таймаутом.
func WaitForCondition(t *testing.T, timeout time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// AssertMessageReceived проверяет, что сервер получил сообщение с определённым opcode.
func AssertMessageReceived(t *testing.T, server *MockServer, opcode int) map[string]any {
	t.Helper()

	messages := server.GetReceivedMessages()
	for _, msg := range messages {
		if opcodeVal, ok := msg["opcode"].(float64); ok && int(opcodeVal) == opcode {
			return msg
		}
	}

	t.Fatalf("message with opcode %d not found in received messages", opcode)
	return nil
}

// AssertNoError проверяет отсутствие ошибки и останавливает тест при её наличии.
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	require.NoError(t, err, msgAndArgs...)
}

// StringPtr возвращает указатель на строку.
func StringPtr(s string) *string {
	return &s
}

// BoolPtr возвращает указатель на bool.
func BoolPtr(b bool) *bool {
	return &b
}

// Int64Ptr возвращает указатель на int64.
func Int64Ptr(i int64) *int64 {
	return &i
}

// IntPtr возвращает указатель на int.
func IntPtr(i int) *int {
	return &i
}

// StartMockServerWithRESTClient создаёт mock сервер и возвращает REST клиент для управления через API.
// REST API автоматически включен при создании сервера.
func StartMockServerWithRESTClient(t *testing.T) (*MockServer, *RESTClient) {
	server := StartMockServer(t)
	client := NewRESTClient("http://" + server.Addr())
	return server, client
}
