package gomax

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fresh-milkshake/gomax/logger"
	"github.com/fresh-milkshake/gomax/mockserver"
	"github.com/fresh-milkshake/gomax/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullFlow_RegisterAndSendMessage тестирует полный flow: регистрация → SYNC → отправка сообщения.
func TestFullFlow_RegisterAndSendMessage(t *testing.T) {
	server := mockserver.StartMockServer(t)

	server.SetHandler(6, func(msg map[string]any) map[string]any {
		return mockserver.SessionInitResponse(0)
	})

	server.SetHandler(17, func(msg map[string]any) map[string]any {
		return mockserver.AuthRequestResponse(0, testTempToken)
	})

	server.SetHandler(18, func(msg map[string]any) map[string]any {
		return mockserver.AuthResponse(0, testRegToken, "")
	})

	server.SetHandler(23, func(msg map[string]any) map[string]any {
		return mockserver.AuthConfirmResponse(0, testAuthToken)
	})

	server.SetHandler(19, func(msg map[string]any) map[string]any {
		profile := map[string]any{
			"contact": map[string]any{
				"id":    int64(123456),
				"phone": testPhone,
				"names": []map[string]any{
					{
						"firstName": testFirstName,
						"lastName":  testLastName,
						"type":      "FIRST_LAST",
					},
				},
			},
		}
		return mockserver.SyncResponse(0, profile, nil)
	})

	server.SetHandler(21, func(msg map[string]any) map[string]any {
		profile := map[string]any{
			"contact": map[string]any{
				"id":    int64(123456),
				"phone": testPhone,
				"names": []map[string]any{
					{
						"firstName": testFirstName,
						"lastName":  testLastName,
						"type":      "FIRST_LAST",
					},
				},
			},
		}
		return mockserver.SyncResponse(0, profile, nil)
	})

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)
		text := message["text"].(string)
		return mockserver.SendMessageResponse(0, testChatID, testMessageID, text)
	})

	workDir := t.TempDir()

	client, err := NewMaxClient(ClientConfig{
		Phone:        testPhone,
		URI:          server.URL(),
		WorkDir:      workDir,
		Registration: true,
		FirstName:    testFirstName,
		LastName:     mockserver.StringPtr(testLastName),
		Logger:       logger.Nop(),
		CodeProvider: func(ctx context.Context) (string, error) {
			return testVerifyCode, nil
		},
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)

	err = client.Start(ctx)
	require.NoError(t, err)

	msg, err := client.SendMessage(ctx, "Hello after registration!", testChatID, true, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, "Hello after registration!", msg.Text)
}

// TestFullFlow_LoginAndSendMessage тестирует полный flow: логин → SYNC → отправка сообщения.
func TestFullFlow_LoginAndSendMessage(t *testing.T) {
	server := mockserver.StartMockServer(t)

	server.SetHandler(6, func(msg map[string]any) map[string]any {
		return mockserver.SessionInitResponse(0)
	})

	server.SetHandler(17, func(msg map[string]any) map[string]any {
		return mockserver.AuthRequestResponse(0, testTempToken)
	})

	server.SetHandler(18, func(msg map[string]any) map[string]any {
		return mockserver.AuthResponse(0, "", testLoginToken)
	})

	server.SetHandler(19, func(msg map[string]any) map[string]any {
		profile := map[string]any{
			"contact": map[string]any{
				"id":    int64(123456),
				"phone": testPhone,
				"names": []map[string]any{
					{
						"firstName": "Logged",
						"lastName":  "User",
						"type":      "FIRST_LAST",
					},
				},
			},
		}
		chats := []map[string]any{
			mockserver.TestChat(testChatID, "DIALOG", ""),
		}
		return mockserver.SyncResponse(0, profile, chats)
	})

	server.SetHandler(21, func(msg map[string]any) map[string]any {
		profile := map[string]any{
			"contact": map[string]any{
				"id":    int64(123456),
				"phone": testPhone,
				"names": []map[string]any{
					{
						"firstName": "Logged",
						"lastName":  "User",
						"type":      "FIRST_LAST",
					},
				},
			},
		}
		chats := []map[string]any{
			mockserver.TestChat(testChatID, "DIALOG", ""),
		}
		return mockserver.SyncResponse(0, profile, chats)
	})

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)
		text := message["text"].(string)
		return mockserver.SendMessageResponse(0, testChatID, testMessageID, text)
	})

	workDir := t.TempDir()

	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Logger:  logger.Nop(),
		CodeProvider: func(ctx context.Context) (string, error) {
			return testVerifyCode, nil
		},
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	msg, err := client.SendMessage(ctx, "Hello after login!", testChatID, true, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, msg)
}

// TestFullFlow_CreateGroupAndSendMessage тестирует создание группы и отправку сообщения.
func TestFullFlow_CreateGroupAndSendMessage(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	groupChatID := int64(999)

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)

		if attaches, ok := message["attaches"].([]interface{}); ok && len(attaches) > 0 {
			attach := attaches[0].(map[string]any)
			if attach["_type"] == "CONTROL" {
				title := attach["title"].(string)
				return mockserver.CreateGroupResponse(0, groupChatID, title, testMessageID)
			}
		}

		text := message["text"].(string)
		chatID := int64(payload["chatId"].(float64))
		return mockserver.SendMessageResponse(0, chatID, testMessageID+1, text)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	chat, createMsg, err := client.CreateGroup(ctx, "Integration Test Group", []int64{testUserID}, true)
	require.NoError(t, err)
	assert.NotNil(t, chat)
	assert.NotNil(t, createMsg)
	require.NotNil(t, chat.Title)
	assert.Equal(t, "Integration Test Group", *chat.Title)

	msg, err := client.SendMessage(ctx, "First message in new group!", groupChatID, true, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, "First message in new group!", msg.Text)
}

// TestFullFlow_AddReactionToMessage тестирует отправку сообщения и добавление реакции.
func TestFullFlow_AddReactionToMessage(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	sentMessageID := int64(0)

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		sentMessageID = testMessageID
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)
		text := message["text"].(string)
		return mockserver.SendMessageResponse(0, testChatID, sentMessageID, text)
	})

	server.SetHandler(178, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		reaction := payload["reaction"].(map[string]any)
		reactionID := reaction["id"].(string)
		return mockserver.AddReactionResponse(0, "12345", reactionID)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	msg, err := client.SendMessage(ctx, "Message with reaction", testChatID, true, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, msg)

	messageIDStr := "12345"
	reactionInfo, err := client.AddReaction(ctx, testChatID, messageIDStr, "👍")
	require.NoError(t, err)
	assert.NotNil(t, reactionInfo)
	assert.Equal(t, 1, reactionInfo.TotalCount)
}

// TestFullFlow_EditAndDeleteMessage тестирует отправку → редактирование → удаление.
func TestFullFlow_EditAndDeleteMessage(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)
		text := message["text"].(string)
		return mockserver.SendMessageResponse(0, testChatID, testMessageID, text)
	})

	server.SetHandler(67, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		text := payload["text"].(string)
		return mockserver.EditMessageResponse(0, testChatID, testMessageID, text)
	})

	server.SetHandler(66, func(msg map[string]any) map[string]any {
		return mockserver.DeleteMessageResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	msg, err := client.SendMessage(ctx, "Original message", testChatID, true, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "Original message", msg.Text)

	editedMsg, err := client.EditMessage(ctx, testChatID, testMessageID, "Edited message", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "Edited message", editedMsg.Text)

	err = client.DeleteMessage(ctx, testChatID, []int64{testMessageID}, false)
	require.NoError(t, err)
}

// TestFullFlow_GroupManagement тестирует полный цикл управления группой.
func TestFullFlow_GroupManagement(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	groupChatID := int64(888)

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)

		if attaches, ok := message["attaches"].([]interface{}); ok && len(attaches) > 0 {
			attach := attaches[0].(map[string]any)
			if attach["_type"] == "CONTROL" {
				title := attach["title"].(string)
				return mockserver.CreateGroupResponse(0, groupChatID, title, testMessageID)
			}
		}
		return mockserver.SendMessageResponse(0, groupChatID, testMessageID, "")
	})

	server.SetHandler(77, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		operation := payload["operation"].(string)
		if operation == "add" {
			return mockserver.InviteUsersResponse(0, groupChatID)
		}
		return mockserver.RemoveUsersResponse(0, groupChatID)
	})

	server.SetHandler(55, func(msg map[string]any) map[string]any {
		return mockserver.ChangeGroupSettingsResponse(0, groupChatID)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	chat, _, err := client.CreateGroup(ctx, "Managed Group", []int64{testUserID}, true)
	require.NoError(t, err)
	require.NotNil(t, chat.Title)
	assert.Equal(t, "Managed Group", *chat.Title)

	err = client.InviteUsersToGroup(ctx, groupChatID, []int64{testUserID + 1, testUserID + 2}, true)
	require.NoError(t, err)

	allCanPin := true
	err = client.ChangeGroupSettings(ctx, groupChatID, &allCanPin, nil, nil, nil, nil)
	require.NoError(t, err)

	newName := "Renamed Group"
	err = client.ChangeGroupProfile(ctx, groupChatID, &newName, nil)
	require.NoError(t, err)

	err = client.RemoveUsersFromGroup(ctx, groupChatID, []int64{testUserID + 2}, 0)
	require.NoError(t, err)
}

// TestFullFlow_ContactManagement тестирует полный цикл работы с контактами.
func TestFullFlow_ContactManagement(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(46, func(msg map[string]any) map[string]any {
		user := mockserver.TestUser(testUserID, "Found", "Contact", "+79995555555")
		return mockserver.SearchByPhoneResponse(0, user)
	})

	server.SetHandler(34, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		action := payload["action"].(string)
		if action == "ADD" {
			contact := mockserver.TestContact(testUserID, "Found", "Contact", "+79995555555")
			return mockserver.AddContactResponse(0, contact)
		}
		return mockserver.RemoveContactResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	user, err := client.SearchByPhone(ctx, "+79995555555")
	require.NoError(t, err)
	require.NotEmpty(t, user.Names)
	assert.Equal(t, "Found", *user.Names[0].FirstName)

	contact, err := client.AddContact(ctx, testUserID)
	require.NoError(t, err)
	assert.NotNil(t, contact)

	err = client.RemoveContact(ctx, testUserID)
	require.NoError(t, err)
}

// TestFullFlow_MessageWithNotifications тестирует отправку сообщения и получение уведомлений.
func TestFullFlow_MessageWithNotifications(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)
		text := message["text"].(string)
		return mockserver.SendMessageResponse(0, testChatID, testMessageID, text)
	})

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

	msg, err := client.SendMessage(ctx, "Message to trigger notification", testChatID, true, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, msg)

	notification := mockserver.NotifMessageResponse(map[string]any{
		"id":       int64(99999),
		"chatId":   testChatID,
		"senderId": testUserID + 100,
		"text":     "Reply to your message",
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
		assert.Equal(t, "Reply to your message", receivedMessage.Text)
	case <-time.After(5 * time.Second):
		t.Fatal("Notification was not received")
	}
}

// TestFullFlow_FolderOperations тестирует полный цикл работы с папками.
func TestFullFlow_FolderOperations(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	createdFolderID := ""

	server.SetHandler(274, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		title := payload["title"].(string)
		id := payload["id"].(string)
		createdFolderID = id
		return mockserver.CreateFolderResponse(0, id, title)
	})

	server.SetHandler(272, func(msg map[string]any) map[string]any {
		folders := []map[string]any{
			mockserver.TestFolder(createdFolderID, "Test Folder", []int64{testChatID}),
		}
		return mockserver.GetFoldersResponse(0, folders, 1)
	})

	server.SetHandler(276, func(msg map[string]any) map[string]any {
		return mockserver.DeleteFolderResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	folderUpdate, err := client.CreateFolder(ctx, "Test Folder", []int64{testChatID}, nil)
	require.NoError(t, err)
	assert.NotNil(t, folderUpdate)

	folderList, err := client.GetFolders(ctx, 0)
	require.NoError(t, err)
	assert.NotNil(t, folderList)

	if createdFolderID != "" {
		_, err = client.DeleteFolder(ctx, createdFolderID)
		require.NoError(t, err)
	}
}

// TestFullFlow_ChatHistory тестирует получение истории и работу с сообщениями.
func TestFullFlow_ChatHistory(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(49, func(msg map[string]any) map[string]any {
		messages := []map[string]any{
			mockserver.TestMessage(1, testChatID, testUserID, "Message 1"),
			mockserver.TestMessage(2, testChatID, testUserID, "Message 2"),
			mockserver.TestMessage(3, testChatID, testUserID, "Message 3"),
		}
		return mockserver.FetchHistoryResponse(0, messages)
	})

	server.SetHandler(55, func(msg map[string]any) map[string]any {
		return mockserver.PinMessageResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	messages, err := client.FetchHistory(ctx, testChatID, nil, 10, 10)
	require.NoError(t, err)
	assert.Len(t, messages, 3)

	err = client.PinMessage(ctx, testChatID, messages[0].ID, true)
	require.NoError(t, err)
}

// TestFullFlow_MultipleChats тестирует работу с несколькими чатами одновременно.
func TestFullFlow_MultipleChats(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	chatIDs := []int64{100, 200, 300}

	server.SetHandler(48, func(msg map[string]any) map[string]any {
		chats := []map[string]any{
			mockserver.TestChat(100, "DIALOG", "Dialog"),
			mockserver.TestChat(200, "CHAT", "Group Chat"),
			mockserver.TestChat(300, "CHANNEL", "Channel"),
		}
		return mockserver.GetChatsResponse(0, chats)
	})

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		chatID := int64(payload["chatId"].(float64))
		message := payload["message"].(map[string]any)
		text := message["text"].(string)
		return mockserver.SendMessageResponse(0, chatID, testMessageID, text)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	chats, err := client.GetChats(ctx, chatIDs)
	require.NoError(t, err)
	assert.Len(t, chats, 3)

	for _, chatID := range chatIDs {
		msg, err := client.SendMessage(ctx, "Message to chat", chatID, true, nil, nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, msg)
	}
}

// TestFullFlow_ErrorRecovery тестирует восстановление после ошибок.
func TestFullFlow_ErrorRecovery(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	callCount := 0

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		callCount++
		if callCount == 1 {
			return mockserver.ErrorResponse(0, 64, "TEMPORARY_ERROR", "Temporary error")
		}
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)
		text := message["text"].(string)
		return mockserver.SendMessageResponse(0, testChatID, testMessageID, text)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	_, err := client.SendMessage(ctx, "First attempt", testChatID, true, nil, nil, nil)
	assert.Error(t, err)

	msg, err := client.SendMessage(ctx, "Second attempt", testChatID, true, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, msg)
}

// TestFullFlow_HighVolume тестирует работу с большим количеством операций.
func TestFullFlow_HighVolume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high volume test in short mode")
	}

	server := mockserver.StartMockServerWithDefaults(t)

	messageCounter := int64(0)
	var mu sync.Mutex

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		mu.Lock()
		messageCounter++
		id := messageCounter
		mu.Unlock()

		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)
		text := message["text"].(string)
		return mockserver.SendMessageResponse(0, testChatID, id, text)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	numMessages := 20
	var wg sync.WaitGroup

	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			msg, err := client.SendMessage(ctx, "High volume message", testChatID, true, nil, nil, nil)
			if err != nil {
				t.Logf("Message %d failed: %v", idx, err)
				return
			}
			assert.NotNil(t, msg)
		}(i)
	}

	wg.Wait()

	mu.Lock()
	finalCount := messageCounter
	mu.Unlock()

	assert.GreaterOrEqual(t, finalCount, int64(1), "At least one message should be sent")
}
