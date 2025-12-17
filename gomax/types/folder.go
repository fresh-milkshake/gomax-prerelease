package types

// Папка (фильтр для чатов пользователя).
type Folder struct {
	SourceID   int64   `json:"sourceId"`
	Include    []int64 `json:"include"`
	Options    []any   `json:"options"`
	UpdateTime int64   `json:"updateTime"`
	ID         string  `json:"id"`
	Filters    []any   `json:"filters"`
	Title      string  `json:"title"`
}

// Результат операций с папками.
type FolderUpdate struct {
	FolderOrder []string `json:"folderOrder,omitempty"`
	Folder      *Folder  `json:"folder,omitempty"`
	FolderSync  int64    `json:"folderSync"`
}

// Список папок пользователя.
type FolderList struct {
	FoldersOrder            []string `json:"foldersOrder"`
	Folders                 []Folder `json:"folders"`
	FolderSync              int64    `json:"folderSync"`
	AllFilterExcludeFolders []any    `json:"allFilterExcludeFolders,omitempty"`
}
