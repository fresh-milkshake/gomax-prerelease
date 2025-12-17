package main

import "time"

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

func SyncResponse(seq int, profile map[string]any, chats []map[string]any) map[string]any {
	if profile == nil {
		profile = map[string]any{
			"contact": map[string]any{
				"id":            int64(123456),
				"phone":         "+79991234567",
				"accountStatus": 0,
				"updateTime":    time.Now().UnixMilli(),
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

func NotifAttachResponse(fileID, videoID int64) map[string]any {
	payload := map[string]any{}
	if fileID > 0 {
		payload["fileId"] = fileID
	}
	if videoID > 0 {
		payload["videoId"] = videoID
	}

	return map[string]any{
		"ver":     ProtocolVersion,
		"cmd":     ProtocolCommand,
		"seq":     0,
		"opcode":  OpcodeNotifAttach,
		"payload": payload,
	}
}

func NotifMessageResponse(message map[string]any) map[string]any {
	return map[string]any{
		"ver":     ProtocolVersion,
		"cmd":     ProtocolCommand,
		"seq":     0,
		"opcode":  OpcodeNotifMessage,
		"payload": message,
	}
}

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
