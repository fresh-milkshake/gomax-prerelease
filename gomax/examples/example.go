package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fresh-milkshake/gomax"
	"github.com/fresh-milkshake/gomax/logger"
	"github.com/fresh-milkshake/gomax/types"

	"github.com/charmbracelet/log"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: example <phone_number>")
	}
	phone := os.Args[1]

	customLogger := logger.New(log.DebugLevel)
	client, err := gomax.NewMaxClient(gomax.ClientConfig{
		Phone:   phone,
		WorkDir: "cache",
		Logger:  customLogger,

		CodeProvider: func(ctx context.Context) (string, error) {
			fmt.Print("Enter verification code: ")
			reader := bufio.NewReader(os.Stdin)
			code, err := reader.ReadString('\n')
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(code), nil
		},
	})
	if err != nil {
		log.Fatal("Failed to create client", "err", err)
	}
	defer client.Close()

	client.OnStart(func(ctx context.Context) {
		log.Info("Client started successfully!")

		me := client.Profile()
		if me != nil && len(me.Names) > 0 {
			name := me.Names[0]
			log.Info("Profile",
				"firstName", name.FirstName,
				"lastName", name.LastName,
				"phone", me.Phone,
			)
		}

		chats := client.ChatList()
		log.Info("Chats", "count", len(chats))
		for i, chat := range chats {
			if i >= 5 {
				log.Info("...", "more", len(chats)-5)
				break
			}
			title := "Unknown"
			if chat.Title != nil {
				title = *chat.Title
			}
			log.Info("  Chat", "id", chat.ID, "type", chat.Type, "title", title)
		}

		dialogs := client.DialogList()
		if len(dialogs) > 0 {
			log.Info("Dialogs", "count", len(dialogs))
		}

		channels := client.ChannelList()
		if len(channels) > 0 {
			log.Info("Channels", "count", len(channels))
		}
	})

	client.OnMessage(func(ctx context.Context, msg *types.Message) {
		log.Info("New message received",
			"chatID", msg.ChatID,
			"messageID", msg.ID,
			"text", msg.Text,
		)
	}, nil)

	client.OnMessageEdit(func(ctx context.Context, msg *types.Message) {
		log.Info("Message edited",
			"chatID", msg.ChatID,
			"messageID", msg.ID,
			"text", msg.Text,
		)
	}, nil)

	client.OnMessageDelete(func(ctx context.Context, msg *types.Message) {
		log.Info("Message deleted",
			"chatID", msg.ChatID,
			"messageID", msg.ID,
		)
	}, nil)

	client.OnChatUpdate(func(ctx context.Context, chat *types.Chat) {
		title := "Unknown"
		if chat.Title != nil {
			title = *chat.Title
		}
		log.Info("Chat updated",
			"chatID", chat.ID,
			"type", chat.Type,
			"title", title,
		)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("Starting client...")
	if err := client.Start(ctx); err != nil {
		log.Fatal("Failed to start client", "err", err)
	}

	log.Info("Client is running. Press Ctrl+C to stop.")

	<-sigChan
	log.Info("Received interrupt signal, shutting down...")

	cancel()
	if err := client.Close(); err != nil {
		log.Error("Error closing client", "err", err)
	}

	log.Info("Client stopped gracefully")
	os.Exit(0)
}
