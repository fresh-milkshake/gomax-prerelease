package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStandaloneServer(t *testing.T) {
	server, err := NewMockServer("localhost", 0)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.HandleWebSocket)
	setupAPI(mux, server)
	server.SetHandlerFunc(mux)

	go func() {
		_ = server.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	baseURL := "http://" + server.Addr()

	t.Run("GET /api/status", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/status")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var status map[string]any
		err = json.NewDecoder(resp.Body).Decode(&status)
		require.NoError(t, err)

		assert.Equal(t, server.Addr(), status["address"])
		assert.Equal(t, server.URL(), status["url"])
		assert.Equal(t, false, status["connected"])
		assert.Equal(t, float64(0), status["connections"])
	})

	t.Run("POST /api/message", func(t *testing.T) {
		req := MessageEventRequest{
			ChatID:    123,
			MessageID: 456,
			Text:      "Hello from test",
			SenderID:  789,
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/api/message", "application/json", bytes.NewBuffer(data))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]any
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "ok", result["status"])
	})

	t.Run("POST /api/relay", func(t *testing.T) {
		relayMsg := map[string]any{
			"ver":    11,
			"cmd":    0,
			"opcode": 128,
			"payload": map[string]any{
				"custom": "data",
			},
		}

		data, err := json.Marshal(relayMsg)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/api/relay", "application/json", bytes.NewBuffer(data))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]any
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "ok", result["status"])
	})

	t.Run("WebSocket connection and broadcast", func(t *testing.T) {
		wsURL := "ws://" + server.Addr() + "/"
		conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn1.Close()

		conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn2.Close()

		time.Sleep(50 * time.Millisecond)

		statusResp, err := http.Get(baseURL + "/api/status")
		require.NoError(t, err)
		defer statusResp.Body.Close()

		var status map[string]any
		err = json.NewDecoder(statusResp.Body).Decode(&status)
		require.NoError(t, err)
		assert.Equal(t, true, status["connected"])
		assert.Equal(t, float64(2), status["connections"])

		req := MessageEventRequest{
			ChatID:    123,
			MessageID: 789,
			Text:      "Broadcast test",
			SenderID:  456,
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/api/message", "application/json", bytes.NewBuffer(data))
		require.NoError(t, err)
		resp.Body.Close()

		var msg1, msg2 map[string]any

		done1 := make(chan bool)
		done2 := make(chan bool)

		go func() {
			err := conn1.ReadJSON(&msg1)
			require.NoError(t, err)
			done1 <- true
		}()

		go func() {
			err := conn2.ReadJSON(&msg2)
			require.NoError(t, err)
			done2 <- true
		}()

		select {
		case <-done1:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for message on conn1")
		}

		select {
		case <-done2:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for message on conn2")
		}

		assert.Equal(t, float64(128), msg1["opcode"])
		assert.Equal(t, float64(128), msg2["opcode"])

		payload1, ok := msg1["payload"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "Broadcast test", payload1["text"])
	})

	t.Run("POST /api/message_edit", func(t *testing.T) {
		req := MessageEditEventRequest{
			ChatID:    123,
			MessageID: 456,
			Text:      "Edited text",
			SenderID:  789,
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/api/message_edit", "application/json", bytes.NewBuffer(data))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("POST /api/reaction", func(t *testing.T) {
		req := ReactionEventRequest{
			ChatID:    123,
			MessageID: "456",
			Reaction:  "👍",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/api/reaction", "application/json", bytes.NewBuffer(data))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/messages", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/messages")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]any
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		messages, ok := result["messages"].([]any)
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(messages), 0)
	})

	t.Run("DELETE /api/messages", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", baseURL+"/api/messages", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp2, err := http.Get(baseURL + "/api/messages")
		require.NoError(t, err)
		defer resp2.Body.Close()

		var result map[string]any
		err = json.NewDecoder(resp2.Body).Decode(&result)
		require.NoError(t, err)

		messages, ok := result["messages"].([]any)
		require.True(t, ok)
		assert.Equal(t, 0, len(messages))
	})

	server.Stop()
}
