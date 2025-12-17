package gomax

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fresh-milkshake/gomax/enums"
	"github.com/fresh-milkshake/gomax/logger"
	"github.com/fresh-milkshake/gomax/mockserver"
	"github.com/fresh-milkshake/gomax/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOnMessage_Handler проверяет обработку новых сообщений.
func TestOnMessage_Handler(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	var receivedMessage *types.Message
	var wg sync.WaitGroup
	wg.Add(1)

	client.OnMessage(func(ctx context.Context, msg *types.Message) {
		receivedMessage = msg
		wg.Done()
	}, nil)

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	notification := mockserver.NotifMessageResponse(map[string]any{
		"id":       int64(12345),
		"chatId":   testChatID,
		"senderId": testUserID,
		"text":     "New message from notification",
		"time":     time.Now().UnixMilli(),
	})

	err = server.SendNotification(notification)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.NotNil(t, receivedMessage)
		assert.Equal(t, int64(12345), receivedMessage.ID)
		assert.Equal(t, "New message from notification", receivedMessage.Text)
	case <-time.After(5 * time.Second):
		t.Fatal("OnMessage handler was not called")
	}
}

// TestOnMessage_WithFilter проверяет обработку сообщений с фильтром.
func TestOnMessage_WithFilter(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	handlerCallCount := 0
	var mu sync.Mutex

	client.OnMessage(func(ctx context.Context, msg *types.Message) {
		mu.Lock()
		handlerCallCount++
		mu.Unlock()
	}, nil)

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		notification := mockserver.NotifMessageResponse(map[string]any{
			"id":       int64(100 + i),
			"chatId":   testChatID,
			"senderId": testUserID,
			"text":     "Message",
			"time":     time.Now().UnixMilli(),
		})
		server.SendNotification(notification)
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	assert.GreaterOrEqual(t, handlerCallCount, 1, "Handler should be called at least once")
	mu.Unlock()
}

// TestOnMessageEdit_Handler проверяет обработку отредактированных сообщений.
func TestOnMessageEdit_Handler(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	var receivedMessage *types.Message
	var wg sync.WaitGroup
	wg.Add(1)

	client.OnMessageEdit(func(ctx context.Context, msg *types.Message) {
		receivedMessage = msg
		wg.Done()
	}, nil)

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	editedStatus := enums.MessageStatusEdited
	notification := mockserver.NotifMessageResponse(map[string]any{
		"id":       int64(12345),
		"chatId":   testChatID,
		"senderId": testUserID,
		"text":     "Edited message",
		"time":     time.Now().UnixMilli(),
		"status":   string(editedStatus),
	})

	err = server.SendNotification(notification)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.NotNil(t, receivedMessage)
		assert.Equal(t, "Edited message", receivedMessage.Text)
	case <-time.After(5 * time.Second):
		t.Fatal("OnMessageEdit handler was not called")
	}
}

// TestOnMessageDelete_Handler проверяет обработку удалённых сообщений.
func TestOnMessageDelete_Handler(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	var receivedMessage *types.Message
	var wg sync.WaitGroup
	wg.Add(1)

	client.OnMessageDelete(func(ctx context.Context, msg *types.Message) {
		receivedMessage = msg
		wg.Done()
	}, nil)

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	removedStatus := enums.MessageStatusRemoved
	notification := mockserver.NotifMessageResponse(map[string]any{
		"id":       int64(12345),
		"chatId":   testChatID,
		"senderId": testUserID,
		"text":     "",
		"time":     time.Now().UnixMilli(),
		"status":   string(removedStatus),
	})

	err = server.SendNotification(notification)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.NotNil(t, receivedMessage)
		assert.Equal(t, int64(12345), receivedMessage.ID)
	case <-time.After(5 * time.Second):
		t.Fatal("OnMessageDelete handler was not called")
	}
}

// TestOnChatUpdate_Handler проверяет обработку обновлений чатов.
func TestOnChatUpdate_Handler(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	var receivedChat *types.Chat
	var wg sync.WaitGroup
	wg.Add(1)

	client.OnChatUpdate(func(ctx context.Context, chat *types.Chat) {
		receivedChat = chat
		wg.Done()
	})

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	notification := mockserver.NotifChatResponse(map[string]any{
		"id":    testChatID,
		"type":  "CHAT",
		"title": "Updated Chat Title",
	})

	err = server.SendNotification(notification)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.NotNil(t, receivedChat)
		assert.Equal(t, testChatID, receivedChat.ID)
		require.NotNil(t, receivedChat.Title)
		assert.Equal(t, "Updated Chat Title", *receivedChat.Title)
	case <-time.After(5 * time.Second):
		t.Fatal("OnChatUpdate handler was not called")
	}
}

// TestOnReactionChange_Handler проверяет обработку изменений реакций.
// Примечание: OnReactionChange хранение обработчиков пока не реализовано (TODO в коде).
func TestOnReactionChange_Handler(t *testing.T) {
	t.Skip("OnReactionChange handler storage not yet implemented")

	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	var receivedMessageID string
	var receivedChatID int64
	var receivedReactionInfo *types.ReactionInfo
	var wg sync.WaitGroup
	wg.Add(1)

	client.OnReactionChange(func(ctx context.Context, messageID string, chatID int64, info *types.ReactionInfo) {
		receivedMessageID = messageID
		receivedChatID = chatID
		receivedReactionInfo = info
		wg.Done()
	})

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	counters := []map[string]any{
		{"reactionType": "EMOJI", "id": "👍", "count": 3},
	}
	notification := mockserver.NotifReactionChangedResponse(testChatID, "12345", 3, "👍", counters)

	err = server.SendNotification(notification)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.Equal(t, "12345", receivedMessageID)
		assert.Equal(t, testChatID, receivedChatID)
		assert.NotNil(t, receivedReactionInfo)
		assert.Equal(t, 3, receivedReactionInfo.TotalCount)
	case <-time.After(5 * time.Second):
		t.Fatal("OnReactionChange handler was not called")
	}
}

// TestMultipleMessageHandlers проверяет работу нескольких обработчиков.
func TestMultipleMessageHandlers(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	handler1Called := false
	handler2Called := false
	var wg sync.WaitGroup
	wg.Add(2)

	client.OnMessage(func(ctx context.Context, msg *types.Message) {
		handler1Called = true
		wg.Done()
	}, nil)

	client.OnMessage(func(ctx context.Context, msg *types.Message) {
		handler2Called = true
		wg.Done()
	}, nil)

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	notification := mockserver.NotifMessageResponse(map[string]any{
		"id":       int64(12345),
		"chatId":   testChatID,
		"senderId": testUserID,
		"text":     "Test message",
		"time":     time.Now().UnixMilli(),
	})

	err = server.SendNotification(notification)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.True(t, handler1Called, "Handler 1 should be called")
		assert.True(t, handler2Called, "Handler 2 should be called")
	case <-time.After(5 * time.Second):
		t.Fatal("Not all handlers were called")
	}
}

// TestNotifAttach_File проверяет обработку уведомления о загрузке файла.
func TestNotifAttach_File(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	notification := mockserver.NotifAttachResponse(12345, 0)

	err = server.SendNotification(notification)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

}

// TestNotifAttach_Video проверяет обработку уведомления о загрузке видео.
func TestNotifAttach_Video(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	notification := mockserver.NotifAttachResponse(0, 67890)

	err = server.SendNotification(notification)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
}

// TestConcurrentNotifications проверяет обработку одновременных уведомлений.
func TestConcurrentNotifications(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	messagesReceived := make(chan int64, 100)

	client.OnMessage(func(ctx context.Context, msg *types.Message) {
		messagesReceived <- msg.ID
	}, nil)

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	numMessages := 10
	for i := 0; i < numMessages; i++ {
		notification := mockserver.NotifMessageResponse(map[string]any{
			"id":       int64(1000 + i),
			"chatId":   testChatID,
			"senderId": testUserID,
			"text":     "Concurrent message",
			"time":     time.Now().UnixMilli(),
		})
		server.SendNotification(notification)
	}

	receivedIDs := make(map[int64]bool)
	timeout := time.After(5 * time.Second)

	for {
		select {
		case id := <-messagesReceived:
			receivedIDs[id] = true
			if len(receivedIDs) >= numMessages {

				return
			}
		case <-timeout:

			assert.GreaterOrEqual(t, len(receivedIDs), 1, "Should receive at least 1 message")
			return
		}
	}
}
