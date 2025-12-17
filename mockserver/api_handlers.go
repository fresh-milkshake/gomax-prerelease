package mockserver

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *MockServer) setupRESTAPI(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/message", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := buildMessageEventFromMap(req)
		if err := s.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/message_edit", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := buildMessageEditEventFromMap(req)
		if err := s.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/message_delete", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := buildMessageDeleteEventFromMap(req)
		if err := s.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/reaction", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := buildReactionEventFromMap(req)
		if err := s.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/chat_update", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := buildChatUpdateEventFromMap(req)
		if err := s.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/contact_update", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := buildContactUpdateEventFromMap(req)
		if err := s.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/typing", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := buildTypingEventFromMap(req)
		if err := s.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/relay", func(w http.ResponseWriter, r *http.Request) {
		var message map[string]any
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		if err := s.Broadcast(message); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("GET /api/messages", func(w http.ResponseWriter, r *http.Request) {
		messages := s.GetReceivedMessages()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"messages": messages})
	})

	mux.HandleFunc("DELETE /api/messages", func(w http.ResponseWriter, r *http.Request) {
		s.ClearReceivedMessages()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"address":     s.Addr(),
			"url":         s.URL(),
			"connected":   s.IsConnected(),
			"connections": s.GetConnectionsCount(),
		})
	})
}
