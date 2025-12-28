package mockserver

import "time"

// SessionInitResponse создаёт ответ на SESSION_INIT.
func SessionInitResponse(seq int) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeSessionInit,
		"payload": map[string]any{
			"status": StatusOK,
		},
	}
}

// AuthRequestResponse создаёт ответ на AUTH_REQUEST с временным токеном.
func AuthRequestResponse(seq int, tempToken string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeAuthRequest,
		"payload": map[string]any{
			"token": tempToken,
		},
	}
}

// AuthResponse создаёт ответ на AUTH (проверка кода).
// Если registerToken не пустой, возвращает токен для регистрации.
// Если loginToken не пустой, возвращает токен для входа.
func AuthResponse(seq int, registerToken, loginToken string) map[string]any {
	tokenAttrs := map[string]any{}

	if registerToken != "" {
		tokenAttrs[TokenTypeRegister] = map[string]any{
			"token": registerToken,
		}
	}

	if loginToken != "" {
		tokenAttrs[TokenTypeLogin] = map[string]any{
			"token": loginToken,
		}
	}

	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeAuth,
		"payload": map[string]any{
			"tokenAttrs": tokenAttrs,
		},
	}
}

// AuthConfirmResponse создаёт ответ на AUTH_CONFIRM (завершение регистрации).
func AuthConfirmResponse(seq int, authToken string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeAuthConfirm,
		"payload": map[string]any{
			"token": authToken,
		},
	}
}

// SyncResponse создаёт ответ на SYNC с профилем и чатами.
func SyncResponse(seq int, profile map[string]any, chats []map[string]any) map[string]any {
	if profile == nil {
		profile = map[string]any{
			"contact": map[string]any{
				"id":    int64(123456),
				"phone": "+79991234567",
				"names": []map[string]any{
					{
						"firstName": "Test",
						"lastName":  "User",
						"type":      NameTypeFirstLast,
					},
				},
			},
		}
	}

	if chats == nil {
		chats = []map[string]any{}
	}

	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeSync,
		"payload": map[string]any{
			"profile": profile,
			"chats":   chats,
		},
	}
}

// SendMessageResponse создаёт ответ на MSG_SEND.
func SendMessageResponse(seq int, chatID, messageID int64, text string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeMsgSend,
		"payload": map[string]any{
			"id":       messageID,
			"chatId":   chatID,
			"text":     text,
			"senderId": int64(123456),
			"time":     time.Now().UnixMilli(),
		},
	}
}

// EditMessageResponse создаёт ответ на MSG_EDIT.
func EditMessageResponse(seq int, chatID, messageID int64, text string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeMsgEdit,
		"payload": map[string]any{
			"id":       messageID,
			"chatId":   chatID,
			"text":     text,
			"senderId": int64(123456),
			"time":     time.Now().UnixMilli(),
			"status":   MessageStatusEdited,
		},
	}
}

// DeleteMessageResponse создаёт ответ на MSG_DELETE.
func DeleteMessageResponse(seq int) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeMsgDelete,
		"payload": map[string]any{
			"status": StatusOK,
		},
	}
}

// FetchHistoryResponse создаёт ответ на CHAT_HISTORY.
func FetchHistoryResponse(seq int, messages []map[string]any) map[string]any {
	if messages == nil {
		messages = []map[string]any{}
	}

	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeChatHistory,
		"payload": map[string]any{
			"messages": messages,
		},
	}
}

// PinMessageResponse создаёт ответ на CHAT_UPDATE (pin).
func PinMessageResponse(seq int) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeChatUpdate,
		"payload": map[string]any{
			"status": StatusOK,
		},
	}
}

// AddReactionResponse создаёт ответ на MSG_REACTION.
func AddReactionResponse(seq int, messageID string, reaction string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeMsgReaction,
		"payload": map[string]any{
			"reactionInfo": map[string]any{
				"totalCount":   1,
				"yourReaction": reaction,
				"counters": []map[string]any{
					{
						"reactionType": ReactionTypeEmoji,
						"id":           reaction,
						"count":        1,
					},
				},
			},
		},
	}
}

// RemoveReactionResponse создаёт ответ на MSG_CANCEL_REACTION.
func RemoveReactionResponse(seq int) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeMsgCancelReaction,
		"payload": map[string]any{
			"reactionInfo": map[string]any{
				"totalCount":   0,
				"yourReaction": nil,
				"counters":     []map[string]any{},
			},
		},
	}
}

// GetQRResponse создаёт ответ на GET_QR.
func GetQRResponse(seq int, trackID string, link string, pollInterval int, expiresAt int64) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeGetQR,
		"payload": map[string]any{
			"trackId":         trackID,
			"qrLink":          link,
			"pollingInterval": pollInterval,
			"expiresAt":       expiresAt,
		},
	}
}

// GetQRStatusResponse создаёт ответ на GET_QR_STATUS.
func GetQRStatusResponse(seq int, loginAvailable bool, expiresAt int64) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeGetQRStatus,
		"payload": map[string]any{
			"status": map[string]any{
				"loginAvailable": loginAvailable,
				"expiresAt":      expiresAt,
			},
		},
	}
}

// LoginByQRResponse создаёт ответ на LOGIN_BY_QR.
func LoginByQRResponse(seq int, loginToken string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeLoginByQR,
		"payload": map[string]any{
			"tokenAttrs": map[string]any{
				TokenTypeLogin: map[string]any{
					"token": loginToken,
				},
			},
		},
	}
}

// GetReactionsResponse создаёт ответ на MSG_GET_REACTIONS.
func GetReactionsResponse(seq int, messagesReactions map[string]any) map[string]any {
	if messagesReactions == nil {
		messagesReactions = map[string]any{}
	}

	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeMsgGetReactions,
		"payload": map[string]any{
			"messagesReactions": messagesReactions,
		},
	}
}

// GetUsersResponse создаёт ответ на CONTACT_INFO.
func GetUsersResponse(seq int, users []map[string]any) map[string]any {
	if users == nil {
		users = []map[string]any{}
	}

	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeContactInfo,
		"payload": map[string]any{
			"contacts": users,
		},
	}
}

// SearchByPhoneResponse создаёт ответ на CONTACT_INFO_BY_PHONE.
func SearchByPhoneResponse(seq int, user map[string]any) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeContactInfoByPhone,
		"payload": map[string]any{
			"contact": user,
		},
	}
}

// AddContactResponse создаёт ответ на CONTACT_UPDATE (add).
func AddContactResponse(seq int, contact map[string]any) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeContactUpdate,
		"payload": map[string]any{
			"contact": contact,
		},
	}
}

// RemoveContactResponse создаёт ответ на CONTACT_UPDATE (remove).
func RemoveContactResponse(seq int) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeContactUpdate,
		"payload": map[string]any{
			"status": StatusOK,
		},
	}
}

// CreateGroupResponse создаёт ответ на MSG_SEND (создание группы).
func CreateGroupResponse(seq int, chatID int64, chatName string, messageID int64) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeMsgSend,
		"payload": map[string]any{
			"id":       messageID,
			"chatId":   chatID,
			"senderId": int64(123456),
			"time":     time.Now().UnixMilli(),
			"chat": map[string]any{
				"id":    chatID,
				"type":  ChatTypeChat,
				"title": chatName,
			},
		},
	}
}

// InviteUsersResponse создаёт ответ на CHAT_MEMBERS_UPDATE (add).
func InviteUsersResponse(seq int, chatID int64) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeChatMembersUpdate,
		"payload": map[string]any{
			"chat": map[string]any{
				"id":   chatID,
				"type": ChatTypeChat,
			},
		},
	}
}

// RemoveUsersResponse создаёт ответ на CHAT_MEMBERS_UPDATE (remove).
func RemoveUsersResponse(seq int, chatID int64) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeChatMembersUpdate,
		"payload": map[string]any{
			"chat": map[string]any{
				"id":   chatID,
				"type": ChatTypeChat,
			},
		},
	}
}

// ChangeGroupSettingsResponse создаёт ответ на CHAT_UPDATE (settings).
func ChangeGroupSettingsResponse(seq int, chatID int64) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeChatUpdate,
		"payload": map[string]any{
			"chat": map[string]any{
				"id":   chatID,
				"type": ChatTypeChat,
			},
		},
	}
}

// JoinGroupResponse создаёт ответ на CHAT_JOIN.
func JoinGroupResponse(seq int, chatID int64, chatName string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeChatJoin,
		"payload": map[string]any{
			"chat": map[string]any{
				"id":    chatID,
				"type":  ChatTypeChat,
				"title": chatName,
			},
		},
	}
}

// ReworkInviteLinkResponse создаёт ответ на CHAT_UPDATE (revoke link).
func ReworkInviteLinkResponse(seq int, chatID int64, newLink string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeChatUpdate,
		"payload": map[string]any{
			"chat": map[string]any{
				"id":          chatID,
				"type":        ChatTypeChat,
				"privateLink": newLink,
			},
		},
	}
}

// GetChatsResponse создаёт ответ на CHAT_INFO.
func GetChatsResponse(seq int, chats []map[string]any) map[string]any {
	if chats == nil {
		chats = []map[string]any{}
	}

	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeChatInfo,
		"payload": map[string]any{
			"chats": chats,
		},
	}
}

// LoadMembersResponse создаёт ответ на CHAT_MEMBERS.
func LoadMembersResponse(seq int, members []map[string]any, marker int) map[string]any {
	if members == nil {
		members = []map[string]any{}
	}

	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeChatMembers,
		"payload": map[string]any{
			"members": members,
			"marker":  marker,
		},
	}
}

// JoinChannelResponse создаёт ответ на CHAT_JOIN (channel).
func JoinChannelResponse(seq int) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeChatJoin,
		"payload": map[string]any{
			"status": StatusOK,
		},
	}
}

// ChangeProfileResponse создаёт ответ на PROFILE.
func ChangeProfileResponse(seq int) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeProfile,
		"payload": map[string]any{
			"status": StatusOK,
		},
	}
}

// CreateFolderResponse создаёт ответ на FOLDERS_UPDATE (create).
func CreateFolderResponse(seq int, folderID string, title string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeFoldersUpdate,
		"payload": map[string]any{
			"folderSync": 1,
			"folders": []map[string]any{
				{
					"id":    folderID,
					"title": title,
				},
			},
		},
	}
}

// GetFoldersResponse создаёт ответ на FOLDERS_GET.
func GetFoldersResponse(seq int, folders []map[string]any, folderSync int) map[string]any {
	if folders == nil {
		folders = []map[string]any{}
	}

	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeFoldersGet,
		"payload": map[string]any{
			"folderSync": folderSync,
			"folders":    folders,
		},
	}
}

// UpdateFolderResponse создаёт ответ на FOLDERS_UPDATE (update).
func UpdateFolderResponse(seq int, folderID string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeFoldersUpdate,
		"payload": map[string]any{
			"folderSync": 2,
		},
	}
}

// DeleteFolderResponse создаёт ответ на FOLDERS_DELETE.
func DeleteFolderResponse(seq int) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeFoldersDelete,
		"payload": map[string]any{
			"folderSync": 3,
		},
	}
}

// GetSessionsResponse создаёт ответ на SESSIONS_INFO.
func GetSessionsResponse(seq int, sessions []map[string]any) map[string]any {
	if sessions == nil {
		sessions = []map[string]any{}
	}

	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeSessionsInfo,
		"payload": map[string]any{
			"sessions": sessions,
		},
	}
}

// PhotoUploadResponse создаёт ответ на PHOTO_UPLOAD.
func PhotoUploadResponse(seq int, uploadURL string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodePhotoUpload,
		"payload": map[string]any{
			"url": uploadURL,
		},
	}
}

// FileUploadResponse создаёт ответ на FILE_UPLOAD.
func FileUploadResponse(seq int, uploadURL string, fileID int64) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeFileUpload,
		"payload": map[string]any{
			"info": []map[string]any{
				{
					"url":    uploadURL,
					"fileId": fileID,
				},
			},
		},
	}
}

// VideoUploadResponse создаёт ответ на VIDEO_UPLOAD.
func VideoUploadResponse(seq int, uploadURL string, videoID int64, token string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeVideoUpload,
		"payload": map[string]any{
			"info": []map[string]any{
				{
					"url":     uploadURL,
					"videoId": videoID,
					"token":   token,
				},
			},
		},
	}
}

// NotifAttachResponse создаёт уведомление NOTIF_ATTACH.
func NotifAttachResponse(fileID, videoID int64) map[string]any {
	payload := map[string]any{}
	if fileID > 0 {
		payload["fileId"] = fileID
	}
	if videoID > 0 {
		payload["videoId"] = videoID
	}

	return map[string]any{
		"ver":     11,
		"cmd":     0,
		"seq":     0,
		"opcode":  OpcodeNotifAttach,
		"payload": payload,
	}
}

// NotifMessageResponse создаёт уведомление NOTIF_MESSAGE.
func NotifMessageResponse(message map[string]any) map[string]any {
	return map[string]any{
		"ver":     11,
		"cmd":     0,
		"seq":     0,
		"opcode":  OpcodeNotifMessage,
		"payload": message,
	}
}

// NotifChatResponse создаёт уведомление NOTIF_CHAT.
func NotifChatResponse(chat map[string]any) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    0,
		"opcode": OpcodeNotifChat,
		"payload": map[string]any{
			"chat": chat,
		},
	}
}

// NotifReactionChangedResponse создаёт уведомление NOTIF_MSG_REACTIONS_CHANGED.
func NotifReactionChangedResponse(chatID int64, messageID string, totalCount int, yourReaction string, counters []map[string]any) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    0,
		"opcode": OpcodeNotifMsgReactionsChanged,
		"payload": map[string]any{
			"chatId":       chatID,
			"messageId":    messageID,
			"totalCount":   totalCount,
			"yourReaction": yourReaction,
			"counters":     counters,
		},
	}
}

// ErrorResponse создаёт ответ с ошибкой.
func ErrorResponse(seq int, opcode int, errorCode string, errorMessage string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": opcode,
		"payload": map[string]any{
			"error":   errorCode,
			"message": errorMessage,
		},
	}
}

// GetVideoByIdResponse создаёт ответ на VIDEO_PLAY.
func GetVideoByIdResponse(seq int, videoURL string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeVideoPlay,
		"payload": map[string]any{
			"url": videoURL,
		},
	}
}

// GetFileByIdResponse создаёт ответ на FILE_DOWNLOAD.
func GetFileByIdResponse(seq int, fileURL string) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeFileDownload,
		"payload": map[string]any{
			"url": fileURL,
		},
	}
}

// ResolveLinkResponse создаёт ответ на LINK_INFO.
func ResolveLinkResponse(seq int) map[string]any {
	return map[string]any{
		"ver":    ProtocolVersion,
		"cmd":    ProtocolCommand,
		"seq":    seq,
		"opcode": OpcodeLinkInfo,
		"payload": map[string]any{
			"status": StatusOK,
		},
	}
}
