package mockserver

const (
	OpcodeSessionInit              = 6
	OpcodeProfile                  = 16
	OpcodeAuthRequest              = 17
	OpcodeAuth                     = 18
	OpcodeLogin                    = 19
	OpcodeLogout                   = 20
	OpcodeSync                     = 21
	OpcodeConfig                   = 22
	OpcodeAuthConfirm              = 23
	OpcodeContactInfo              = 32
	OpcodeContactUpdate            = 34
	OpcodeContactInfoByPhone       = 46
	OpcodeChatInfo                 = 48
	OpcodeChatHistory              = 49
	OpcodeChatUpdate               = 55
	OpcodeChatJoin                 = 57
	OpcodeChatMembers              = 59
	OpcodeChatMembersUpdate        = 77
	OpcodeMsgSend                  = 64
	OpcodeMsgEdit                  = 67
	OpcodeMsgDelete                = 66
	OpcodeMsgReaction              = 178
	OpcodeMsgCancelReaction        = 179
	OpcodeMsgGetReactions          = 180
	OpcodeFoldersGet               = 272
	OpcodeFoldersUpdate            = 274
	OpcodeFoldersDelete            = 276
	OpcodeSessionsInfo             = 96
	OpcodePhotoUpload              = 80
	OpcodeFileUpload               = 87
	OpcodeFileDownload             = 88
	OpcodeVideoUpload              = 82
	OpcodeVideoPlay                = 83
	OpcodeLinkInfo                 = 89
	OpcodeNotifMessage             = 128
	OpcodeNotifTyping              = 129
	OpcodeNotifChat                = 135
	OpcodeNotifAttach              = 136
	OpcodeNotifContact             = 131
	OpcodeNotifMsgDelete           = 142
	OpcodeNotifMsgReactionsChanged = 155
)
