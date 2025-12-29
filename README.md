<p align="center">
    <img src="gomax/assets/logo.svg" alt="GoMax" width="400">
</p>

<p align="center">
    <strong>Go клиент для API мессенджера Max</strong>
</p>

<p align="center">
    <img src="https://img.shields.io/badge/go-1.22+-00ADD8.svg" alt="Go 1.22+">
    <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT">
    <a href="https://deepwiki.com/fresh-milkshake/gomax-prerelease">
        <img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki">
    </a>
</p>

> [!WARNING]
> * Это **неофициальная** библиотека для работы с внутренним API Max.
> * Использование может **нарушать условия предоставления услуг** сервиса.
> * **Вы используете её исключительно на свой страх и риск.**
> * **Разработчики и контрибьюторы не несут никакой ответственности** за любые последствия использования этого пакета, включая, но не ограничиваясь: блокировку аккаунтов, утерю данных, юридические риски и любые другие проблемы.
> * API может быть изменён в любой момент без предупреждения.

## Описание

**GoMax** — Go библиотека для работы с API мессенджера Max. Предоставляет интерфейс для отправки сообщений, управления чатами, каналами и диалогами через WebSocket соединение.

### Основные возможности

- Авторизация по номеру телефона (логин и регистрация)
- Отправка, редактирование и удаление сообщений
- Загрузка фото, файлов и видео
- Управление группами и каналами
- История сообщений
- Работа с реакциями
- Управление папками
- Обработка событий через обработчики

## Установка

> [!IMPORTANT]
> Для работы библиотеки требуется Go 1.22 или выше

```bash
go get github.com/fresh-milkshake/gomax
```

## Быстрый старт

### Минимальный пример

```go
package main

import (
    "context"
    "time"

    "github.com/charmbracelet/log"
    "github.com/fresh-milkshake/gomax"
    "github.com/fresh-milkshake/gomax/types"
)

func main() {
    // Создание клиента
    client, err := gomax.NewMaxClient(gomax.ClientConfig{
        Phone:   "+79991234567",
        WorkDir: "cache", // Директория для хранения сессии
    })
    if err != nil {
        log.Fatal("Failed to create client", "err", err)
    }
    defer client.Close()

    // Обработчик входящих сообщений
    client.OnMessage(func(ctx context.Context, msg *types.Message) {
        log.Info("New message", "text", msg.Text)
    }, nil)

    // Обработчик успешного подключения
    client.OnStart(func(ctx context.Context) {
        log.Info("Client started!")
        
        // Отправка сообщения
        _, err := client.SendMessage(ctx, "Привет!", 12345, true, nil, nil, nil)
        if err != nil {
            log.Error("Failed to send message", "err", err)
        }
    })

    // Запуск клиента
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    if err := client.Start(ctx); err != nil {
        log.Fatal("Failed to start", "err", err)
    }

    // Держим клиент активным
    select {}
}
```

### Использование CodeProvider для автоматизации

```go
client, err := gomax.NewMaxClient(gomax.ClientConfig{
    Phone:   "+79991234567",
    WorkDir: "cache",
    // Автоматическое получение кода (например, из Telegram бота)
    CodeProvider: func(ctx context.Context) (string, error) {
        // Ваша логика получения кода
        return "123456", nil
    },
})
```

### Кастомное логирование

По умолчанию библиотека логирует в stderr с уровнем Info. Вы можете передать свой логгер:

```go
import (
    "io"
    "os"
    
    "github.com/charmbracelet/log"
    "github.com/fresh-milkshake/gomax"
)

// Логгер с уровнем Debug
debugLogger := log.NewWithOptions(os.Stderr, log.Options{
    Level:           log.DebugLevel,
    ReportTimestamp: true,
})

// Логгер с записью в файл
file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
fileLogger := log.NewWithOptions(file, log.Options{
    Level:           log.InfoLevel,
    ReportTimestamp: true,
})

// Логгер в несколько мест (терминал + файл)
multiWriter := io.MultiWriter(os.Stderr, file)
multiLogger := log.NewWithOptions(multiWriter, log.Options{
    Level:           log.DebugLevel,
    ReportTimestamp: true,
})

client, _ := gomax.NewMaxClient(gomax.ClientConfig{
    Phone:   "+79991234567",
    Logger:  multiLogger, // Передаём свой логгер
})
```

## API Reference

### Сообщения

```go
// Отправка текстового сообщения
msg, err := client.SendMessage(ctx, "Текст", chatID, notify, photos, files, replyTo)

// Редактирование сообщения
msg, err := client.EditMessage(ctx, chatID, messageID, "Новый текст", photos, files)

// Удаление сообщений
err := client.DeleteMessage(ctx, chatID, []int64{messageID1, messageID2}, forMe)

// История сообщений
messages, err := client.FetchHistory(ctx, chatID, fromMessageID, forward, backward)
```

### Реакции

```go
// Добавить реакцию
info, err := client.AddReaction(ctx, chatID, messageID, "👍")

// Удалить реакцию
info, err := client.RemoveReaction(ctx, chatID, messageID)

// Получить реакции
reactions, err := client.GetReactions(ctx, chatID, []string{messageID1, messageID2})
```

### Группы и каналы

```go
// Создать группу
chat, msg, err := client.CreateGroup(ctx, "Название", []int64{userID1, userID2}, notify)

// Пригласить пользователей
err := client.InviteUsersToGroup(ctx, chatID, []int64{userID}, showHistory)

// Удалить пользователей
err := client.RemoveUsersFromGroup(ctx, chatID, []int64{userID}, cleanMsgPeriod)

// Изменить настройки группы
err := client.ChangeGroupSettings(ctx, chatID, allCanPinMessage, slowMode, onlyAdminCanAddMember, ...)

// Изменить профиль группы
err := client.ChangeGroupProfile(ctx, chatID, name, description)

// Присоединиться по ссылке
chat, err := client.JoinGroup(ctx, "https://max.ru/join/...")

// Загрузить участников
members, nextMarker, err := client.LoadMembers(ctx, chatID, marker, count)
```

### Пользователи и контакты

```go
// Получить информацию о пользователе
user, err := client.GetUser(ctx, userID)

// Получить информацию о нескольких пользователях
users, err := client.GetUsers(ctx, []int64{userID1, userID2})

// Поиск по телефону
user, err := client.SearchByPhone(ctx, "+79991234567")

// Добавить/удалить контакт
contact, err := client.AddContact(ctx, userID)
err := client.RemoveContact(ctx, userID)
```

### Папки

```go
// Создать папку
folder, err := client.CreateFolder(ctx, "Работа", []int64{chatID1, chatID2}, excludeIDs)

// Получить папки
folders, err := client.GetFolders(ctx, folderSync)

// Обновить папку
folder, err := client.UpdateFolder(ctx, folderID, "Новое название", includeIDs, excludeIDs, removeIDs)

// Удалить папку
result, err := client.DeleteFolder(ctx, folderID)
```

### Обработчики событий

```go
// Новые сообщения
client.OnMessage(func(ctx context.Context, msg *types.Message) {
    log.Info("Message", "text", msg.Text)
}, filter) // filter может быть nil

// Редактирование сообщений
client.OnMessageEdit(func(ctx context.Context, msg *types.Message) {
    log.Info("Edited", "text", msg.Text)
}, nil)

// Удаление сообщений
client.OnMessageDelete(func(ctx context.Context, msg *types.Message) {
    log.Info("Deleted", "id", msg.ID)
}, nil)

// Обновление чата
client.OnChatUpdate(func(ctx context.Context, chat *types.Chat) {
    log.Info("Chat updated", "id", chat.ID)
})

// Изменение реакций
client.OnReactionChange(func(ctx context.Context, messageID string, chatID int64, info *types.ReactionInfo) {
    log.Info("Reaction changed", "messageID", messageID)
})

// Успешный старт
client.OnStart(func(ctx context.Context) {
    log.Info("Started!")
})
```

### Фильтры сообщений

```go
import "github.com/fresh-milkshake/gomax/filters"

// Фильтр по chatID
filter := &filters.Filter{ChatID: &chatID}

// Фильтр по тексту
text := "привет"
filter := &filters.Filter{TextContains: &text}

// Комбинированный фильтр
filter := &filters.Filter{
    ChatID:       &chatID,
    TextContains: &text,
}

client.OnMessage(handler, filter)
```

## Структура проекта

```
gomax/
├── client.go           # Основной клиент MaxClient
├── client_methods.go   # Методы API (сообщения, группы, контакты и т.д.)
├── errors.go           # Определения ошибок
├── constants/          # Константы (URL, таймауты и т.д.)
├── database/           # Хранение сессии (SQLite)
├── enums/              # Перечисления (opcodes, типы сообщений и т.д.)
├── files/              # Работа с файлами для загрузки
├── filters/            # Фильтры сообщений
├── logger/             # Хелперы для логирования
├── payloads/           # Структуры запросов к API
├── types/              # Структуры данных (Message, Chat, User и т.д.)
└── utils/              # Утилиты (JSON, форматирование)
```

## Связанные проекты

- **[PyMax](https://github.com/ink-developer/PyMax)** — Python версия этой библиотеки

## Лицензия

MIT License. См. файл [LICENSE](LICENSE).
