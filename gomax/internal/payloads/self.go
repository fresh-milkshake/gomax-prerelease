package payloads

// Payload для изменения профиля пользователя.
type ChangeProfilePayload struct {
	FirstName   string  `json:"firstName"`
	LastName    *string `json:"lastName,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Payload для создания папки.
type CreateFolderPayload struct {
	ID      string        `json:"id"`
	Title   string        `json:"title"`
	Include []int64       `json:"include"`
	Filters []interface{} `json:"filters"`
}

// Payload для получения папок.
type GetFolderPayload struct {
	FolderSync int `json:"folderSync"`
}

// Payload для обновления папки.
type UpdateFolderPayload struct {
	ID      string        `json:"id"`
	Title   string        `json:"title"`
	Include []int64       `json:"include"`
	Filters []interface{} `json:"filters"`
	Options []interface{} `json:"options"`
}

// Payload для удаления папки.
type DeleteFolderPayload struct {
	FolderIDs []string `json:"folderIds"`
}
