package mockserver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRESTClientIntegration(t *testing.T) {
	server := StartMockServer(t)

	baseURL := "http://" + server.Addr()
	client := NewRESTClient(baseURL)

	time.Sleep(50 * time.Millisecond)

	t.Run("TriggerMessage via REST", func(t *testing.T) {
		event := MessageEvent{
			ChatID:    123,
			MessageID: 456,
			Text:      "Hello from REST",
			SenderID:  789,
		}

		err := client.TriggerMessage(event)
		require.NoError(t, err)
	})

	t.Run("TriggerMessageEdit via REST", func(t *testing.T) {
		event := MessageEditEvent{
			ChatID:    123,
			MessageID: 456,
			Text:      "Edited text",
			SenderID:  789,
		}

		err := client.TriggerMessageEdit(event)
		require.NoError(t, err)
	})

	t.Run("TriggerReaction via REST", func(t *testing.T) {
		event := ReactionEvent{
			ChatID:    123,
			MessageID: "456",
			Reaction:  "👍",
		}

		err := client.TriggerReaction(event)
		require.NoError(t, err)
	})

	t.Run("Relay via REST", func(t *testing.T) {
		message := map[string]any{
			"ver":    11,
			"cmd":    0,
			"opcode": 128,
			"payload": map[string]any{
				"custom": "relay data",
			},
		}

		err := client.Relay(message)
		require.NoError(t, err)
	})

	t.Run("GetMessages via REST", func(t *testing.T) {
		messages, err := client.GetMessages()
		require.NoError(t, err)
		assert.NotNil(t, messages)
	})

	t.Run("ClearMessages via REST", func(t *testing.T) {
		err := client.ClearMessages()
		require.NoError(t, err)

		messages, err := client.GetMessages()
		require.NoError(t, err)
		assert.Equal(t, 0, len(messages))
	})

	t.Run("GetStatus via REST", func(t *testing.T) {
		status, err := client.GetStatus()
		require.NoError(t, err)
		assert.Equal(t, server.Addr(), status.Address)
		assert.Equal(t, server.URL(), status.URL)
	})
}

func TestStartMockServerWithRESTClient(t *testing.T) {
	server, client := StartMockServerWithRESTClient(t)

	require.NotNil(t, server)
	require.NotNil(t, client)

	event := MessageEvent{
		ChatID:    123,
		MessageID: 456,
		Text:      "Test message",
		SenderID:  789,
	}

	err := client.TriggerMessage(event)
	require.NoError(t, err)

	status, err := client.GetStatus()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, status.Connections, 0)
}
