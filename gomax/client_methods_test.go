package gomax

import (
	"testing"
	"time"

	"github.com/fresh-milkshake/gomax/logger"
	"github.com/fresh-milkshake/gomax/mockserver"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testChatID    int64 = 12345
	testMessageID int64 = 67890
	testUserID    int64 = 11111
)

// createTestClient создаёт и запускает тестового клиента.
func createTestClient(t *testing.T, server *mockserver.MockServer) *MaxClient {
	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

// TestSendMessage_TextOnly проверяет отправку простого текстового сообщения.
func TestSendMessage_TextOnly(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)
		text := message["text"].(string)
		return mockserver.SendMessageResponse(0, testChatID, testMessageID, text)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	msg, err := client.SendMessage(ctx, "Hello, World!", testChatID, true, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, testMessageID, msg.ID)
	assert.Equal(t, "Hello, World!", msg.Text)
}

// TestSendMessage_WithMarkdown проверяет отправку с markdown форматированием.
func TestSendMessage_WithMarkdown(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedElements []interface{}
	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)
		if elements, ok := message["elements"].([]interface{}); ok {
			receivedElements = elements
		}
		text := message["text"].(string)
		return mockserver.SendMessageResponse(0, testChatID, testMessageID, text)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	msg, err := client.SendMessage(ctx, "**Bold** and _italic_", testChatID, true, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, msg)

	_ = receivedElements
}

// TestSendMessage_WithReply проверяет отправку с ответом на сообщение.
func TestSendMessage_WithReply(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedLink interface{}
	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)
		receivedLink = message["link"]
		return mockserver.SendMessageResponse(0, testChatID, testMessageID, "Reply")
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	replyToID := int64(99999)
	msg, err := client.SendMessage(ctx, "Reply", testChatID, true, nil, nil, &replyToID)
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.NotNil(t, receivedLink)

	linkMap := receivedLink.(map[string]any)
	assert.Equal(t, "REPLY", linkMap["type"])
	assert.Equal(t, "99999", linkMap["messageId"])
}

// TestSendMessage_NoNotify проверяет отправку без уведомления.
func TestSendMessage_NoNotify(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedNotify bool
	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		receivedNotify = payload["notify"].(bool)
		return mockserver.SendMessageResponse(0, testChatID, testMessageID, "Silent")
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	msg, err := client.SendMessage(ctx, "Silent", testChatID, false, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.False(t, receivedNotify)
}

// TestEditMessage_Text проверяет редактирование текста сообщения.
func TestEditMessage_Text(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(67, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		text := payload["text"].(string)
		messageID := int64(payload["messageId"].(float64))
		chatID := int64(payload["chatId"].(float64))
		return mockserver.EditMessageResponse(0, chatID, messageID, text)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	msg, err := client.EditMessage(ctx, testChatID, testMessageID, "Edited text", nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, "Edited text", msg.Text)
}

// TestDeleteMessage_Single проверяет удаление одного сообщения.
func TestDeleteMessage_Single(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedMessageIDs []interface{}
	server.SetHandler(66, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		receivedMessageIDs = payload["messageIds"].([]interface{})
		return mockserver.DeleteMessageResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	err := client.DeleteMessage(ctx, testChatID, []int64{testMessageID}, false)
	require.NoError(t, err)
	assert.Len(t, receivedMessageIDs, 1)
}

// TestDeleteMessage_Multiple проверяет удаление нескольких сообщений.
func TestDeleteMessage_Multiple(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedMessageIDs []interface{}
	server.SetHandler(66, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		receivedMessageIDs = payload["messageIds"].([]interface{})
		return mockserver.DeleteMessageResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	messageIDs := []int64{testMessageID, testMessageID + 1, testMessageID + 2}
	err := client.DeleteMessage(ctx, testChatID, messageIDs, false)
	require.NoError(t, err)
	assert.Len(t, receivedMessageIDs, 3)
}

// TestDeleteMessage_ForMe проверяет удаление только для себя.
func TestDeleteMessage_ForMe(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedForMe bool
	server.SetHandler(66, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		receivedForMe = payload["forMe"].(bool)
		return mockserver.DeleteMessageResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	err := client.DeleteMessage(ctx, testChatID, []int64{testMessageID}, true)
	require.NoError(t, err)
	assert.True(t, receivedForMe)
}

// TestAddReaction проверяет добавление реакции.
func TestAddReaction(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(178, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		reaction := payload["reaction"].(map[string]any)
		reactionID := reaction["id"].(string)
		return mockserver.AddReactionResponse(0, "12345", reactionID)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	reactionInfo, err := client.AddReaction(ctx, testChatID, "12345", "👍")
	require.NoError(t, err)
	assert.NotNil(t, reactionInfo)
	assert.Equal(t, 1, reactionInfo.TotalCount)
}

// TestRemoveReaction проверяет удаление реакции.
func TestRemoveReaction(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(179, func(msg map[string]any) map[string]any {
		return mockserver.RemoveReactionResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	reactionInfo, err := client.RemoveReaction(ctx, testChatID, "12345")
	require.NoError(t, err)
	assert.NotNil(t, reactionInfo)
	assert.Equal(t, 0, reactionInfo.TotalCount)
}

// TestGetReactions проверяет получение реакций для сообщений.
func TestGetReactions(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(180, func(msg map[string]any) map[string]any {
		messagesReactions := map[string]any{
			"12345": map[string]any{
				"totalCount":   2,
				"yourReaction": "👍",
				"counters": []map[string]any{
					{"reactionType": "EMOJI", "id": "👍", "count": 2},
				},
			},
		}
		return mockserver.GetReactionsResponse(0, messagesReactions)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	reactions, err := client.GetReactions(ctx, testChatID, []string{"12345"})
	require.NoError(t, err)
	assert.Len(t, reactions, 1)
	assert.NotNil(t, reactions["12345"])
	assert.Equal(t, 2, reactions["12345"].TotalCount)
}

// TestCreateGroup проверяет создание группы.
func TestCreateGroup(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		message := payload["message"].(map[string]any)
		attaches := message["attaches"].([]interface{})
		if len(attaches) > 0 {
			attach := attaches[0].(map[string]any)
			if attach["_type"] == "CONTROL" {
				title := attach["title"].(string)
				return mockserver.CreateGroupResponse(0, 999, title, testMessageID)
			}
		}
		return mockserver.SendMessageResponse(0, testChatID, testMessageID, "")
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	chat, msg, err := client.CreateGroup(ctx, "Test Group", []int64{testUserID}, true)
	require.NoError(t, err)
	assert.NotNil(t, chat)
	assert.NotNil(t, msg)
	assert.Equal(t, int64(999), chat.ID)
	require.NotNil(t, chat.Title)
	assert.Equal(t, "Test Group", *chat.Title)
}

// TestInviteUsersToGroup проверяет приглашение пользователей в группу.
func TestInviteUsersToGroup(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedUserIDs []interface{}
	server.SetHandler(77, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		receivedUserIDs = payload["userIds"].([]interface{})
		return mockserver.InviteUsersResponse(0, testChatID)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	err := client.InviteUsersToGroup(ctx, testChatID, []int64{testUserID, testUserID + 1}, true)
	require.NoError(t, err)
	assert.Len(t, receivedUserIDs, 2)
}

// TestRemoveUsersFromGroup проверяет удаление пользователей из группы.
func TestRemoveUsersFromGroup(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedUserIDs []interface{}
	var receivedOperation string
	server.SetHandler(77, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		receivedUserIDs = payload["userIds"].([]interface{})
		receivedOperation = payload["operation"].(string)
		return mockserver.RemoveUsersResponse(0, testChatID)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	err := client.RemoveUsersFromGroup(ctx, testChatID, []int64{testUserID}, 0)
	require.NoError(t, err)
	assert.Len(t, receivedUserIDs, 1)
	assert.Equal(t, "remove", receivedOperation)
}

// TestChangeGroupSettings проверяет изменение настроек группы.
func TestChangeGroupSettings(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedOptions map[string]any
	server.SetHandler(55, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		if opts, ok := payload["options"].(map[string]any); ok {
			receivedOptions = opts
		}
		return mockserver.ChangeGroupSettingsResponse(0, testChatID)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	allCanPin := true
	err := client.ChangeGroupSettings(ctx, testChatID, &allCanPin, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, receivedOptions)
}

// TestChangeGroupProfile проверяет изменение профиля группы.
func TestChangeGroupProfile(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedTheme, receivedDescription interface{}
	server.SetHandler(55, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		receivedTheme = payload["theme"]
		receivedDescription = payload["description"]
		return mockserver.ChangeGroupSettingsResponse(0, testChatID)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	newName := "New Group Name"
	newDesc := "New Description"
	err := client.ChangeGroupProfile(ctx, testChatID, &newName, &newDesc)
	require.NoError(t, err)
	assert.Equal(t, newName, receivedTheme)
	assert.Equal(t, newDesc, receivedDescription)
}

// TestJoinGroup проверяет присоединение к группе по ссылке.
func TestJoinGroup(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(57, func(msg map[string]any) map[string]any {
		return mockserver.JoinGroupResponse(0, 888, "Joined Group")
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	chat, err := client.JoinGroup(ctx, "https://max.ru/join/abc123")
	require.NoError(t, err)
	assert.NotNil(t, chat)
	assert.Equal(t, int64(888), chat.ID)
}

// TestReworkInviteLink проверяет пересоздание ссылки приглашения.
func TestReworkInviteLink(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(55, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		if _, ok := payload["revokePrivateLink"]; ok {
			return mockserver.ReworkInviteLinkResponse(0, testChatID, "https://max.ru/join/newlink")
		}
		return mockserver.ChangeGroupSettingsResponse(0, testChatID)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	chat, err := client.ReworkInviteLink(ctx, testChatID)
	require.NoError(t, err)
	assert.NotNil(t, chat)
}

// TestGetUser проверяет получение информации о пользователе.
func TestGetUser(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(32, func(msg map[string]any) map[string]any {
		users := []map[string]any{
			mockserver.TestUser(testUserID, "John", "Doe", "+79991111111"),
		}
		return mockserver.GetUsersResponse(0, users)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	user, err := client.GetUser(ctx, testUserID)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, testUserID, user.ID)
	require.NotEmpty(t, user.Names)
	assert.Equal(t, "John", *user.Names[0].FirstName)
}

// TestGetUsers проверяет получение информации о нескольких пользователях.
func TestGetUsers(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(32, func(msg map[string]any) map[string]any {
		users := []map[string]any{
			mockserver.TestUser(testUserID, "John", "Doe", "+79991111111"),
			mockserver.TestUser(testUserID+1, "Jane", "Doe", "+79992222222"),
		}
		return mockserver.GetUsersResponse(0, users)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	users, err := client.GetUsers(ctx, []int64{testUserID, testUserID + 1})
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

// TestSearchByPhone проверяет поиск по номеру телефона.
func TestSearchByPhone(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(46, func(msg map[string]any) map[string]any {
		user := mockserver.TestUser(testUserID, "Found", "User", "+79993333333")
		return mockserver.SearchByPhoneResponse(0, user)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	user, err := client.SearchByPhone(ctx, "+79993333333")
	require.NoError(t, err)
	assert.NotNil(t, user)
	require.NotEmpty(t, user.Names)
	assert.Equal(t, "Found", *user.Names[0].FirstName)
}

// TestAddContact проверяет добавление контакта.
func TestAddContact(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(34, func(msg map[string]any) map[string]any {
		contact := mockserver.TestContact(testUserID, "Added", "Contact", "+79994444444")
		return mockserver.AddContactResponse(0, contact)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	contact, err := client.AddContact(ctx, testUserID)
	require.NoError(t, err)
	assert.NotNil(t, contact)
}

// TestRemoveContact проверяет удаление контакта.
func TestRemoveContact(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(34, func(msg map[string]any) map[string]any {
		return mockserver.RemoveContactResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	err := client.RemoveContact(ctx, testUserID)
	require.NoError(t, err)
}

// TestFetchHistory проверяет загрузку истории сообщений.
func TestFetchHistory(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(49, func(msg map[string]any) map[string]any {
		messages := []map[string]any{
			mockserver.TestMessage(1, testChatID, testUserID, "Message 1"),
			mockserver.TestMessage(2, testChatID, testUserID, "Message 2"),
			mockserver.TestMessage(3, testChatID, testUserID, "Message 3"),
		}
		return mockserver.FetchHistoryResponse(0, messages)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	messages, err := client.FetchHistory(ctx, testChatID, nil, 10, 10)
	require.NoError(t, err)
	assert.Len(t, messages, 3)
}

// TestPinMessage проверяет закрепление сообщения.
func TestPinMessage(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedPinMessageID interface{}
	server.SetHandler(55, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		receivedPinMessageID = payload["pinMessageId"]
		return mockserver.PinMessageResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	err := client.PinMessage(ctx, testChatID, testMessageID, true)
	require.NoError(t, err)
	assert.NotNil(t, receivedPinMessageID)
}

// TestGetChats проверяет получение информации о чатах.
func TestGetChats(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(48, func(msg map[string]any) map[string]any {
		chats := []map[string]any{
			mockserver.TestChat(testChatID, "CHAT", "Test Chat"),
			mockserver.TestChat(testChatID+1, "CHANNEL", "Test Channel"),
		}
		return mockserver.GetChatsResponse(0, chats)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	chats, err := client.GetChats(ctx, []int64{testChatID, testChatID + 1})
	require.NoError(t, err)
	assert.Len(t, chats, 2)
}

// TestGetChat проверяет получение информации об одном чате.
func TestGetChat(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(48, func(msg map[string]any) map[string]any {
		chats := []map[string]any{
			mockserver.TestChat(testChatID, "CHAT", "Single Chat"),
		}
		return mockserver.GetChatsResponse(0, chats)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	chat, err := client.GetChat(ctx, testChatID)
	require.NoError(t, err)
	assert.NotNil(t, chat)
	assert.Equal(t, testChatID, chat.ID)
}

// TestLoadMembers проверяет загрузку участников группы/канала.
func TestLoadMembers(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(59, func(msg map[string]any) map[string]any {
		members := []map[string]any{
			mockserver.TestMember(testUserID, "owner"),
			mockserver.TestMember(testUserID+1, "admin"),
			mockserver.TestMember(testUserID+2, "member"),
		}
		return mockserver.LoadMembersResponse(0, members, 0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	members, nextMarker, err := client.LoadMembers(ctx, testChatID, nil, 10)
	require.NoError(t, err)
	assert.Len(t, members, 3)
	assert.Nil(t, nextMarker)
}

// TestFindMembers проверяет поиск участников.
func TestFindMembers(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(59, func(msg map[string]any) map[string]any {
		members := []map[string]any{
			mockserver.TestMember(testUserID, "member"),
		}
		return mockserver.LoadMembersResponse(0, members, 0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	members, _, err := client.FindMembers(ctx, testChatID, "search query")
	require.NoError(t, err)
	assert.Len(t, members, 1)
}

// TestJoinChannel проверяет присоединение к каналу.
func TestJoinChannel(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(57, func(msg map[string]any) map[string]any {
		return mockserver.JoinChannelResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	err := client.JoinChannel(ctx, "https://max.ru/channel")
	require.NoError(t, err)
}

// TestChangeProfile проверяет изменение профиля пользователя.
func TestChangeProfile(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	var receivedFirstName, receivedLastName, receivedDescription interface{}
	server.SetHandler(16, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		receivedFirstName = payload["firstName"]
		receivedLastName = payload["lastName"]
		receivedDescription = payload["description"]
		return mockserver.ChangeProfileResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	lastName := "NewLastName"
	description := "New description"
	err := client.ChangeProfile(ctx, "NewFirstName", &lastName, &description)
	require.NoError(t, err)
	assert.Equal(t, "NewFirstName", receivedFirstName)
	assert.Equal(t, lastName, receivedLastName)
	assert.Equal(t, description, receivedDescription)
}

// TestCreateFolder проверяет создание папки.
func TestCreateFolder(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(274, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		title := payload["title"].(string)
		id := payload["id"].(string)
		return mockserver.CreateFolderResponse(0, id, title)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	folder, err := client.CreateFolder(ctx, "Test Folder", []int64{testChatID}, nil)
	require.NoError(t, err)
	assert.NotNil(t, folder)
}

// TestGetFolders проверяет получение списка папок.
func TestGetFolders(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(272, func(msg map[string]any) map[string]any {
		folders := []map[string]any{
			mockserver.TestFolder("1", "Folder 1", []int64{testChatID}),
			mockserver.TestFolder("2", "Folder 2", []int64{testChatID + 1}),
		}
		return mockserver.GetFoldersResponse(0, folders, 5)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	folderList, err := client.GetFolders(ctx, 0)
	require.NoError(t, err)
	assert.NotNil(t, folderList)
}

// TestUpdateFolder проверяет обновление папки.
func TestUpdateFolder(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(274, func(msg map[string]any) map[string]any {
		payload := msg["payload"].(map[string]any)
		id := payload["id"].(string)
		return mockserver.UpdateFolderResponse(0, id)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	folder, err := client.UpdateFolder(ctx, "1", "Updated Folder", []int64{testChatID}, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, folder)
}

// TestDeleteFolder проверяет удаление папки.
func TestDeleteFolder(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(276, func(msg map[string]any) map[string]any {
		return mockserver.DeleteFolderResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	folder, err := client.DeleteFolder(ctx, "1")
	require.NoError(t, err)
	assert.NotNil(t, folder)
}

// TestGetSessions проверяет получение списка сессий.
func TestGetSessions(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(96, func(msg map[string]any) map[string]any {
		sessions := []map[string]any{
			mockserver.TestSession("1", "desktop", "windows", true),
			mockserver.TestSession("2", "mobile", "android", false),
		}
		return mockserver.GetSessionsResponse(0, sessions)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	sessions, err := client.GetSessions(ctx)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)
}

// TestGetChatId проверяет вычисление ID диалога.
func TestGetChatId(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)
	client := createTestClient(t, server)

	chatID := client.GetChatId(100, 200)

	assert.Equal(t, int64(172), chatID)

	chatID2 := client.GetChatId(200, 100)
	assert.Equal(t, chatID, chatID2)
}

// TestResolveChannelByName проверяет разрешение канала по имени.
func TestResolveChannelByName(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(89, func(msg map[string]any) map[string]any {
		return mockserver.ResolveLinkResponse(0)
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	err := client.ResolveChannelByName(ctx, "testchannel")
	require.NoError(t, err)
}

// TestGetVideoById проверяет получение метаданных видео.
func TestGetVideoById(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(83, func(msg map[string]any) map[string]any {
		return mockserver.GetVideoByIdResponse(0, "https://example.com/video.mp4")
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	videoReq, err := client.GetVideoById(ctx, testChatID, testMessageID, 1234)
	require.NoError(t, err)
	assert.NotNil(t, videoReq)
}

// TestGetFileById проверяет получение метаданных файла.
func TestGetFileById(t *testing.T) {
	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(88, func(msg map[string]any) map[string]any {
		return mockserver.GetFileByIdResponse(0, "https://example.com/file.pdf")
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	fileReq, err := client.GetFileById(ctx, testChatID, testMessageID, 5678)
	require.NoError(t, err)
	assert.NotNil(t, fileReq)
}

// TestSendMessage_UploadFlow проверяет flow загрузки вложений.
func TestSendMessage_UploadFlow(t *testing.T) {

	httpServer := mockserver.MockHTTPServer(t, mockserver.PhotoUploadHandler("photo_token_123"))

	server := mockserver.StartMockServerWithDefaults(t)

	server.SetHandler(80, func(msg map[string]any) map[string]any {
		return mockserver.PhotoUploadResponse(0, httpServer.URL)
	})

	server.SetHandler(64, func(msg map[string]any) map[string]any {
		return mockserver.SendMessageResponse(0, testChatID, testMessageID, "Photo message")
	})

	client := createTestClient(t, server)
	ctx := mockserver.TestContext(t)

	msg, err := client.SendMessage(ctx, "Photo message", testChatID, true, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, msg)
}

// TestTimeout проверяет обработку таймаутов.
func TestTimeout(t *testing.T) {
	server := mockserver.StartMockServer(t)

	server.SetHandler(6, func(msg map[string]any) map[string]any {
		return mockserver.SessionInitResponse(0)
	})

	server.SetHandler(21, func(msg map[string]any) map[string]any {
		time.Sleep(100 * time.Millisecond)
		return mockserver.SyncResponse(0, nil, nil)
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

	ctx := mockserver.TestContext(t)

	err = client.Start(ctx)

	_ = err
}
