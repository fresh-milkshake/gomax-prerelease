package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type RequestHandler func(msg map[string]any) map[string]any

type MockServer struct {
	listener net.Listener
	server   *http.Server
	upgrader websocket.Upgrader

	mu          sync.RWMutex
	connections map[string]*websocket.Conn
	writeMu     sync.Mutex
	handlers    map[int]RequestHandler

	ReceivedMessages []map[string]any
	msgMu            sync.Mutex
}

func NewMockServer(host string, port int) (*MockServer, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	s := &MockServer{
		listener: listener,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		connections:      make(map[string]*websocket.Conn),
		handlers:         make(map[int]RequestHandler),
		ReceivedMessages: make([]map[string]any, 0),
	}

	s.server = &http.Server{}

	return s, nil
}

func (s *MockServer) Start() error {
	return s.server.Serve(s.listener)
}

func (s *MockServer) SetHandlerFunc(handler http.Handler) {
	s.server.Handler = handler
}

func (s *MockServer) Stop() error {
	s.mu.Lock()
	for _, conn := range s.connections {
		conn.Close()
	}
	s.connections = make(map[string]*websocket.Conn)
	s.mu.Unlock()
	return s.server.Close()
}

func (s *MockServer) URL() string {
	return fmt.Sprintf("ws://%s/", s.listener.Addr().String())
}

func (s *MockServer) Addr() string {
	return s.listener.Addr().String()
}

func (s *MockServer) SetHandler(opcode int, handler RequestHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[opcode] = handler
}

func (s *MockServer) GetHandler(opcode int) (RequestHandler, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	handler, ok := s.handlers[opcode]
	return handler, ok
}

func (s *MockServer) DeleteHandler(opcode int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.handlers, opcode)
}

func (s *MockServer) GetAllHandlers() map[int]RequestHandler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[int]RequestHandler)
	for k, v := range s.handlers {
		result[k] = v
	}
	return result
}

func (s *MockServer) Broadcast(message map[string]any) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	s.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(s.connections))
	for _, conn := range s.connections {
		connections = append(connections, conn)
	}
	s.mu.RUnlock()

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	var lastErr error
	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return fmt.Errorf("failed to broadcast to some clients: %w", lastErr)
	}

	return nil
}

func (s *MockServer) GetReceivedMessages() []map[string]any {
	s.msgMu.Lock()
	defer s.msgMu.Unlock()
	result := make([]map[string]any, len(s.ReceivedMessages))
	copy(result, s.ReceivedMessages)
	return result
}

func (s *MockServer) ClearReceivedMessages() {
	s.msgMu.Lock()
	defer s.msgMu.Unlock()
	s.ReceivedMessages = make([]map[string]any, 0)
}

func (s *MockServer) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connections) > 0
}

func (s *MockServer) GetConnectionsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connections)
}

func (s *MockServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	connID := fmt.Sprintf("%p", conn)
	s.mu.Lock()
	s.connections[connID] = conn
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.connections, connID)
		s.mu.Unlock()
		conn.Close()
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg map[string]any
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		s.msgMu.Lock()
		s.ReceivedMessages = append(s.ReceivedMessages, msg)
		s.msgMu.Unlock()

		opcodeVal, _ := msg["opcode"].(float64)
		opcode := int(opcodeVal)
		seqVal, _ := msg["seq"].(float64)
		seq := int(seqVal)

		s.mu.RLock()
		handler, ok := s.handlers[opcode]
		s.mu.RUnlock()

		if ok && handler != nil {
			response := handler(msg)
			if response != nil {
				response["seq"] = seq
				respData, err := json.Marshal(response)
				if err != nil {
					continue
				}
				s.writeMu.Lock()
				conn.WriteMessage(websocket.TextMessage, respData)
				s.writeMu.Unlock()
			}
		}
	}
}
