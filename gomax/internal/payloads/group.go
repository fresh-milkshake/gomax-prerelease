package payloads

// Вложение для создания группы.
type CreateGroupAttach struct {
	Type     string  `json:"_type"`
	Event    string  `json:"event"`
	ChatType string  `json:"chatType"`
	Title    string  `json:"title"`
	UserIDs  []int64 `json:"userIds"`
}

// Сообщение для создания группы.
type CreateGroupMessage struct {
	CID      int64               `json:"cid"`
	Attaches []CreateGroupAttach `json:"attaches"`
}

// Payload для создания группы.
type CreateGroupPayload struct {
	Message CreateGroupMessage `json:"message"`
	Notify  bool               `json:"notify"`
}

// Payload для приглашения пользователей в группу.
type InviteUsersPayload struct {
	ChatID      int64   `json:"chatId"`
	UserIDs     []int64 `json:"userIds"`
	ShowHistory bool    `json:"showHistory"`
	Operation   string  `json:"operation"`
}

// Payload для удаления пользователей из группы.
type RemoveUsersPayload struct {
	ChatID         int64   `json:"chatId"`
	UserIDs        []int64 `json:"userIds"`
	Operation      string  `json:"operation"`
	CleanMsgPeriod int     `json:"cleanMsgPeriod"`
}

// Опции настроек группы.
type ChangeGroupSettingsOptions struct {
	OnlyOwnerCanChangeIconTitle *bool `json:"ONLY_OWNER_CAN_CHANGE_ICON_TITLE,omitempty"`
	AllCanPinMessage            *bool `json:"ALL_CAN_PIN_MESSAGE,omitempty"`
	OnlyAdminCanAddMember       *bool `json:"ONLY_ADMIN_CAN_ADD_MEMBER,omitempty"`
	OnlyAdminCanCall            *bool `json:"ONLY_ADMIN_CAN_CALL,omitempty"`
	MembersCanSeePrivateLink    *bool `json:"MEMBERS_CAN_SEE_PRIVATE_LINK,omitempty"`
}

// Payload для изменения настроек группы.
type ChangeGroupSettingsPayload struct {
	ChatID  int64                      `json:"chatId"`
	Options ChangeGroupSettingsOptions `json:"options"`
}

// Payload для изменения профиля группы.
type ChangeGroupProfilePayload struct {
	ChatID      int64   `json:"chatId"`
	Theme       *string `json:"theme,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Payload для пересоздания ссылки приглашения.
type ReworkInviteLinkPayload struct {
	RevokePrivateLink bool  `json:"revokePrivateLink"`
	ChatID            int64 `json:"chatId"`
}

// Payload для получения информации о чатах.
type GetChatInfoPayload struct {
	ChatIDs []int64 `json:"chatIds"`
}
