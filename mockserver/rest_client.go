package mockserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

type RESTClient struct {
	baseURL string
	client  *http.Client
}

func NewRESTClient(baseURL string) *RESTClient {
	return &RESTClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

type MessageEvent struct {
	ChatID    int64  `json:"chatId"`
	MessageID int64  `json:"messageId"`
	Text      string `json:"text"`
	SenderID  int64  `json:"senderId"`
	Time      int64  `json:"time,omitempty"`
}

type MessageEditEvent struct {
	ChatID    int64  `json:"chatId"`
	MessageID int64  `json:"messageId"`
	Text      string `json:"text"`
	SenderID  int64  `json:"senderId"`
	Time      int64  `json:"time,omitempty"`
}

type MessageDeleteEvent struct {
	ChatID     int64   `json:"chatId"`
	MessageID  int64   `json:"messageId"`
	MessageIDs []int64 `json:"messageIds,omitempty"`
}

type ReactionEvent struct {
	ChatID       int64            `json:"chatId"`
	MessageID    string           `json:"messageId"`
	Reaction     string           `json:"reaction,omitempty"`
	TotalCount   int              `json:"totalCount,omitempty"`
	YourReaction string           `json:"yourReaction,omitempty"`
	Counters     []map[string]any `json:"counters,omitempty"`
}

type ChatUpdateEvent struct {
	ChatID int64          `json:"chatId"`
	Chat   map[string]any `json:"chat,omitempty"`
	Data   map[string]any `json:"data,omitempty"`
}

type ContactUpdateEvent struct {
	Contact map[string]any `json:"contact,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

type TypingEvent struct {
	ChatID int64 `json:"chatId"`
	UserID int64 `json:"userId"`
	Typing bool  `json:"typing"`
}

func (c *RESTClient) TriggerMessage(event MessageEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/message", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *RESTClient) TriggerMessageEdit(event MessageEditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/message_edit", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *RESTClient) TriggerMessageDelete(event MessageDeleteEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/message_delete", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *RESTClient) TriggerReaction(event ReactionEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/reaction", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *RESTClient) TriggerChatUpdate(event ChatUpdateEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/chat_update", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *RESTClient) TriggerContactUpdate(event ContactUpdateEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/contact_update", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *RESTClient) TriggerTyping(event TypingEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/typing", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *RESTClient) Relay(message map[string]any) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/relay", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *RESTClient) GetMessages() ([]map[string]any, error) {
	resp, err := c.client.Get(c.baseURL + "/api/messages")
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Messages []map[string]any `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Messages, nil
}

func (c *RESTClient) ClearMessages() error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/messages", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

type StatusResponse struct {
	Address     string `json:"address"`
	URL         string `json:"url"`
	Connected   bool   `json:"connected"`
	Connections int    `json:"connections"`
}

func (c *RESTClient) GetStatus() (*StatusResponse, error) {
	resp, err := c.client.Get(c.baseURL + "/api/status")
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var status StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}

func StartMockServerWithREST(t *testing.T, port int) (*RESTClient, func()) {
	cmd := exec.Command("go", "run", "../../cmd/mockserver/main.go", "--port", strconv.Itoa(port), "--host", "localhost")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	client := NewRESTClient(fmt.Sprintf("http://localhost:%d", port))

	cleanup := func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}

	return client, cleanup
}
