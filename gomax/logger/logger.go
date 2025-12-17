// Package logger предоставляет интеграцию с библиотекой charmbracelet/log для gomax.
//
// По умолчанию логи выводятся в терминал (os.Stderr) с цветным форматированием.
// Пользователь может передать свой *log.Logger для кастомной настройки логирования,
// включая запись в файл или другие назначения.
//
// Пример кастомного логгера с записью в файл:
//
//	import (
//		"os"
//		"github.com/charmbracelet/log"
//	)
//
//	file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
//	logger := log.NewWithOptions(file, log.Options{
//		Level:           log.DebugLevel,
//		ReportTimestamp: true,
//	})
//
//	client, _ := gomax.NewMaxClient(gomax.ClientConfig{
//		Phone:  "+1234567890",
//		Logger: logger,
//	})
package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

// Default возвращает логгер по умолчанию для gomax.
// Выводит логи в stderr с цветным форматированием и уровнем Info.
func Default() *log.Logger {
	return log.NewWithOptions(os.Stderr, log.Options{
		Level:           log.InfoLevel,
		ReportTimestamp: true,
	})
}

// New создаёт новый логгер с указанным уровнем логирования.
// Выводит логи в stderr с цветным форматированием.
func New(level log.Level) *log.Logger {
	return log.NewWithOptions(os.Stderr, log.Options{
		Level:           level,
		ReportTimestamp: true,
	})
}

// NewWithPrefix создаёт новый логгер с указанным уровнем и префиксом.
func NewWithPrefix(level log.Level, prefix string) *log.Logger {
	return log.NewWithOptions(os.Stderr, log.Options{
		Level:           level,
		ReportTimestamp: true,
		Prefix:          prefix,
	})
}

// Nop возвращает логгер, который ничего не выводит.
// Полезно для тестов или когда логирование не нужно.
func Nop() *log.Logger {
	return log.NewWithOptions(os.Stderr, log.Options{
		Level: log.FatalLevel + 1,
	})
}
