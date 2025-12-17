package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fresh-milkshake/gomax"
	"github.com/fresh-milkshake/gomax/logger"

	"github.com/charmbracelet/log"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: dialogs_example <phone_number>")
	}
	phone := os.Args[1]

	customLogger := logger.New(log.InfoLevel)
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

		dialogs := client.DialogList()
		log.Info("Found dialogs", "count", len(dialogs))

		if len(dialogs) == 0 {
			log.Info("No dialogs found")
			return
		}

		for i, dialog := range dialogs {
			log.Info("",
				"separator", "=",
				"dialog", fmt.Sprintf("%d/%d", i+1, len(dialogs)),
			)
			log.Info("Dialog",
				"id", dialog.ID,
				"lastEventTime", dialog.LastEventTime,
			)

			var fromTime *int64
			if dialog.LastEventTime > 0 {
				fromTime = &dialog.LastEventTime
			} else {
				now := time.Now().UnixMilli()
				fromTime = &now
			}

			messages, err := client.FetchHistory(ctx, dialog.ID, fromTime, 0, 5)
			if err != nil {
				log.Error("Failed to fetch messages",
					"dialogID", dialog.ID,
					"err", err,
				)
				continue
			}

			log.Info("Messages", "count", len(messages))
			if len(messages) == 0 {
				log.Info("  (no messages)")
				continue
			}

			for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
				messages[i], messages[j] = messages[j], messages[i]
			}

			for j, msg := range messages {
				msgTime := time.Unix(msg.Time/1000, 0).Format("2006-01-02 15:04:05")
				text := msg.Text
				if text == "" {
					text = "(no text)"
				}
				log.Info("  Message",
					"#", j+1,
					"id", msg.ID,
					"time", msgTime,
					"text", text,
				)
			}
		}

		log.Info("", "separator", "=")
		log.Info("All dialogs processed", "total", len(dialogs))
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
