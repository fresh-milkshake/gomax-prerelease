package files

import (
	"io"
	"os"
	"path/filepath"
)

// Описывает минимальный интерфейс файла, который может быть загружен в Max.
type BaseFile interface {
	Read() ([]byte, error)
	FileName() string
}

// Представляет локальный или удалённый файл, подготовленный к загрузке.
type File struct {
	path     string
	url      string
	fileName string
}

// Создаёт File из указанного пути к локальному файлу.
func NewFileFromPath(path string) (*File, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return &File{
		path:     abs,
		fileName: filepath.Base(abs),
	}, nil
}

// Создаёт File, ссылающийся на удалённый ресурс по URL.
func NewFileFromURL(url string) *File {
	return &File{
		url:      url,
		fileName: filepath.Base(url),
	}
}

// Читает содержимое файла либо с диска, либо (в дальнейшем) по URL.
func (f *File) Read() ([]byte, error) {
	if f.path != "" {
		return os.ReadFile(f.path)
	}
	return nil, io.ErrUnexpectedEOF
}

// Возвращает имя файла, используемое при загрузке в Max.
func (f *File) FileName() string {
	return f.fileName
}

// Представляет фото‑файл, который может быть проверен и загружен в Max.
type Photo struct {
	*File
}

// Создаёт Photo из локального пути к файлу изображения.
func NewPhotoFromPath(path string) (*Photo, error) {
	f, err := NewFileFromPath(path)
	if err != nil {
		return nil, err
	}
	return &Photo{File: f}, nil
}

// Создаёт Photo из URL удалённого изображения.
func NewPhotoFromURL(url string) *Photo {
	return &Photo{File: NewFileFromURL(url)}
}

// Проверяет, что расширение фото поддерживается Max,
// и возвращает расширение без точки и соответствующий MIME‑тип.
func (p *Photo) ValidatePhoto() (string, string, error) {
	ext := filepath.Ext(p.fileName)
	allowed := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
	}
	mime, ok := allowed[ext]
	if !ok {
		return "", "", &InvalidPhotoExtensionError{Ext: ext}
	}
	return ext[1:], mime, nil
}

// Представляет видео‑файл, подготовленный к загрузке в Max.
type Video struct {
	*File
}

// Создаёт Video из локального пути к файлу.
func NewVideoFromPath(path string) (*Video, error) {
	f, err := NewFileFromPath(path)
	if err != nil {
		return nil, err
	}
	return &Video{File: f}, nil
}

// Создаёт Video из URL удалённого видео.
func NewVideoFromURL(url string) *Video {
	return &Video{File: NewFileFromURL(url)}
}

// Сигнализирует о недопустимом расширении файла изображения.
type InvalidPhotoExtensionError struct {
	Ext string
}

func (e *InvalidPhotoExtensionError) Error() string {
	return "invalid photo extension: " + e.Ext
}
