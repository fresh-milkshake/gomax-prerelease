// Package mockserver предоставляет mock WebSocket сервер для тестирования MaxClient.
package mockserver

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// RequestHandler обрабатывает входящее сообщение и возвращает ответ.
// Если handler возвращает nil, ответ не отправляется.
type RequestHandler func(msg map[string]any) map[string]any

// NotificationSender позволяет отправлять уведомления клиенту.
type NotificationSender interface {
	SendNotification(msg map[string]any) error
}

// MockServer представляет mock WebSocket сервер для тестирования.
type MockServer struct {
	listener net.Listener
	server   *http.Server
	upgrader websocket.Upgrader

	mu          sync.RWMutex
	connections map[string]*websocket.Conn
	writeMu     sync.Mutex
	handlers    map[int]RequestHandler

	notifCh chan map[string]any

	ReceivedMessages []map[string]any
	msgMu            sync.Mutex
}

// NewMockServer создаёт новый mock сервер на случайном порту.
func NewMockServer() (*MockServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
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
		notifCh:          make(chan map[string]any, 100),
		ReceivedMessages: make([]map[string]any, 0),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleWebSocket)

	s.server = &http.Server{
		Handler: mux,
	}

	s.setupRESTAPI(mux)

	return s, nil
}

// Start запускает сервер в фоновом режиме.
func (s *MockServer) Start() {
	go s.server.Serve(s.listener)
	go s.notificationLoop()
}

// notificationLoop отправляет уведомления клиентам (для обратной совместимости).
func (s *MockServer) notificationLoop() {
	for msg := range s.notifCh {
		_ = s.Broadcast(msg)
	}
}

// Stop останавливает сервер.
func (s *MockServer) Stop() {
	close(s.notifCh)
	s.mu.Lock()
	for _, conn := range s.connections {
		conn.Close()
	}
	s.connections = make(map[string]*websocket.Conn)
	s.mu.Unlock()
	s.server.Close()
	s.listener.Close()
}

// URL возвращает WebSocket URL для подключения.
func (s *MockServer) URL() string {
	return fmt.Sprintf("ws://%s/", s.listener.Addr().String())
}

// Addr возвращает адрес сервера.
func (s *MockServer) Addr() string {
	return s.listener.Addr().String()
}

// SetHandler устанавливает обработчик для определённого opcode.
func (s *MockServer) SetHandler(opcode int, handler RequestHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[opcode] = handler
}

// SendNotification отправляет уведомление клиенту (для обратной совместимости).
func (s *MockServer) SendNotification(msg map[string]any) error {
	return s.Broadcast(msg)
}

// Broadcast отправляет сообщение всем подключенным клиентам.
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

// GetConnectionsCount возвращает количество подключенных клиентов.
func (s *MockServer) GetConnectionsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connections)
}

// GetReceivedMessages возвращает все полученные сообщения.
func (s *MockServer) GetReceivedMessages() []map[string]any {
	s.msgMu.Lock()
	defer s.msgMu.Unlock()
	result := make([]map[string]any, len(s.ReceivedMessages))
	copy(result, s.ReceivedMessages)
	return result
}

// ClearReceivedMessages очищает список полученных сообщений.
func (s *MockServer) ClearReceivedMessages() {
	s.msgMu.Lock()
	defer s.msgMu.Unlock()
	s.ReceivedMessages = make([]map[string]any, 0)
}

// IsConnected возвращает true, если есть подключенные клиенты.
func (s *MockServer) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connections) > 0
}

// CloseAllConnections закрывает все активные соединения, но оставляет сервер работающим.
func (s *MockServer) CloseAllConnections() {
	s.mu.Lock()
	connections := make([]*websocket.Conn, 0, len(s.connections))
	for _, conn := range s.connections {
		connections = append(connections, conn)
	}
	s.connections = make(map[string]*websocket.Conn)
	s.mu.Unlock()

	for _, conn := range connections {
		conn.Close()
	}
}

func (s *MockServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
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

// DefaultHandlers устанавливает стандартные обработчики для базовых операций.
func (s *MockServer) DefaultHandlers() {
	s.SetHandler(6, func(msg map[string]any) map[string]any {
		return SessionInitResponse(0)
	})

	s.SetHandler(19, func(msg map[string]any) map[string]any {
		return SyncResponse(0, nil, nil)
	})

	s.SetHandler(21, func(msg map[string]any) map[string]any {
		return SyncResponse(0, nil, nil)
	})
}

// SetupAuthHandlers настраивает обработчики для аутентификации.
func (s *MockServer) SetupAuthHandlers(tempToken, registerToken, loginToken, authToken string) {
	s.SetHandler(17, func(msg map[string]any) map[string]any {
		return AuthRequestResponse(0, tempToken)
	})

	s.SetHandler(18, func(msg map[string]any) map[string]any {
		return AuthResponse(0, registerToken, loginToken)
	})

	s.SetHandler(23, func(msg map[string]any) map[string]any {
		return AuthConfirmResponse(0, authToken)
	})
}
