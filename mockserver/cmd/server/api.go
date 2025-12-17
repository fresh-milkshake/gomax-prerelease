package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func setupAPI(mux *http.ServeMux, server *MockServer) {
	mux.HandleFunc("POST /api/message", func(w http.ResponseWriter, r *http.Request) {
		var req MessageEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := BuildMessageEvent(req)
		if err := server.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/message_edit", func(w http.ResponseWriter, r *http.Request) {
		var req MessageEditEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := BuildMessageEditEvent(req)
		if err := server.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/message_delete", func(w http.ResponseWriter, r *http.Request) {
		var req MessageDeleteEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := BuildMessageDeleteEvent(req)
		if err := server.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/reaction", func(w http.ResponseWriter, r *http.Request) {
		var req ReactionEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := BuildReactionEvent(req)
		if err := server.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/chat_update", func(w http.ResponseWriter, r *http.Request) {
		var req ChatUpdateEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := BuildChatUpdateEvent(req)
		if err := server.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/contact_update", func(w http.ResponseWriter, r *http.Request) {
		var req ContactUpdateEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := BuildContactUpdateEvent(req)
		if err := server.Broadcast(event); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/typing", func(w http.ResponseWriter, r *http.Request) {
		var req TypingEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		event := BuildTypingEvent(req)
		if err := server.Broadcast(event); err != nil {
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

		if err := server.Broadcast(message); err != nil {
			http.Error(w, fmt.Sprintf("failed to broadcast: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("GET /api/messages", func(w http.ResponseWriter, r *http.Request) {
		messages := server.GetReceivedMessages()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"messages": messages})
	})

	mux.HandleFunc("DELETE /api/messages", func(w http.ResponseWriter, r *http.Request) {
		server.ClearReceivedMessages()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"address":     server.Addr(),
			"url":         server.URL(),
			"connected":   server.IsConnected(),
			"connections": server.GetConnectionsCount(),
		})
	})

	mux.HandleFunc("POST /api/handler/{opcode}", func(w http.ResponseWriter, r *http.Request) {
		opcodeStr := r.PathValue("opcode")
		var opcode int
		if _, err := fmt.Sscanf(opcodeStr, "%d", &opcode); err != nil {
			http.Error(w, fmt.Sprintf("invalid opcode: %v", err), http.StatusBadRequest)
			return
		}

		var req struct {
			Response map[string]any `json:"response"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		server.SetHandler(opcode, func(msg map[string]any) map[string]any {
			return req.Response
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("POST /api/default_handlers", func(w http.ResponseWriter, r *http.Request) {
		server.SetHandler(OpcodeSessionInit, func(msg map[string]any) map[string]any {
			seqVal, _ := msg["seq"].(float64)
			return SessionInitResponse(int(seqVal))
		})

		server.SetHandler(OpcodeLogin, func(msg map[string]any) map[string]any {
			seqVal, _ := msg["seq"].(float64)
			return SyncResponse(int(seqVal), nil, nil)
		})

		server.SetHandler(OpcodeSync, func(msg map[string]any) map[string]any {
			seqVal, _ := msg["seq"].(float64)
			return SyncResponse(int(seqVal), nil, nil)
		})

		server.SetHandler(OpcodeAuthRequest, func(msg map[string]any) map[string]any {
			seqVal, _ := msg["seq"].(float64)
			return AuthRequestResponse(int(seqVal), "temp_token_123")
		})

		server.SetHandler(OpcodeAuth, func(msg map[string]any) map[string]any {
			seqVal, _ := msg["seq"].(float64)
			return AuthResponse(int(seqVal), "", "login_token_789")
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})
}
