package gomax

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fresh-milkshake/gomax/enums"
	"github.com/fresh-milkshake/gomax/filters"
	"github.com/fresh-milkshake/gomax/internal/constants"
	"github.com/fresh-milkshake/gomax/internal/database"
	"github.com/fresh-milkshake/gomax/internal/utils"
	"github.com/fresh-milkshake/gomax/logger"
	"github.com/fresh-milkshake/gomax/types"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Описывает обработчик сообщения Max с опциональным фильтром.
type messageHandler struct {
	handler func(context.Context, *types.Message)
	filter  *filters.Filter
}

// Задаёт параметры подключения MaxClient к WebSocket API Max и поведение клиента.
type ClientConfig struct {
	Phone             string
	URI               string
	Host              string
	Port              int
	WorkDir           string
	Token             string
	Reconnect         bool
	ReconnectDelay    time.Duration
	SendFakeTelemetry bool
	Registration      bool
	FirstName         string
	LastName          *string
	MaxRetries        int
	RetryInitialDelay time.Duration
	RetryMaxDelay     time.Duration

	// CodeProvider предоставляет код подтверждения из SMS/звонка.
	// Если не указан, MaxClient запросит код у пользователя через stdin.
	CodeProvider func(ctx context.Context) (string, error)

	// Logger позволяет передать кастомный логгер *log.Logger.
	// Если не указан, используется логгер по умолчанию (stderr, InfoLevel).
	// Для записи в файл или другие назначения создайте свой логгер:
	//
	//   file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	//   logger := log.NewWithOptions(file, log.Options{Level: log.DebugLevel})
	Logger *log.Logger
}

// Предоставляет высокоуровневый доступ к неофициальному WebSocket API мессенджера Max.
// Экземпляр отвечает за подключение, авторизацию, синхронизацию состояния и обработку событий.
type MaxClient struct {
	cfg ClientConfig

	logger     *log.Logger
	db         *database.DB
	httpClient *http.Client
	deviceID   uuid.UUID
	token      string

	connMu      sync.RWMutex
	ws          *websocket.Conn
	isConnected bool

	reconnectMu  sync.Mutex
	reconnecting bool

	seqMu sync.Mutex
	seq   int

	pendingMu sync.Mutex
	pending   map[int]chan map[string]any

	incoming chan map[string]any
	outgoing chan map[string]any

	bgWG     sync.WaitGroup
	bgCancel context.CancelFunc

	stateMu  sync.RWMutex
	Me       *types.Me
	Chats    []types.Chat
	Dialogs  []types.Dialog
	Channels []types.Channel

	onStartHandlers         []func(context.Context)
	onMessageHandlers       []messageHandler
	onMessageEditHandlers   []messageHandler
	onMessageDeleteHandlers []messageHandler
	onChatUpdate            []func(context.Context, *types.Chat)
	onReactionChange        []func(context.Context, string, int64, *types.ReactionInfo)

	fileUploadWaitersMu sync.Mutex
	fileUploadWaiters   map[int64]chan map[string]any
}

// Создаёт новый экземпляр MaxClient с указанной конфигурацией
// и инициализирует локальное хранилище сессии и токена.
func NewMaxClient(cfg ClientConfig) (*MaxClient, error) {
	if cfg.URI == "" {
		cfg.URI = constants.WebsocketURI
	}
	if cfg.Host == "" {
		cfg.Host = constants.Host
	}
	if cfg.Port == 0 {
		cfg.Port = constants.Port
	}
	if cfg.WorkDir == "" {
		cfg.WorkDir = "."
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = constants.DefaultMaxRetries
	}
	if cfg.RetryInitialDelay == 0 {
		cfg.RetryInitialDelay = time.Duration(constants.DefaultRetryInitialDelay * float64(time.Second))
	}
	if cfg.RetryMaxDelay == 0 {
		cfg.RetryMaxDelay = time.Duration(constants.DefaultRetryMaxDelay * float64(time.Second))
	}
	if cfg.ReconnectDelay == 0 {
		cfg.ReconnectDelay = 5 * time.Second
	}
	if !constants.PhoneRegex.MatchString(cfg.Phone) {
		return nil, &InvalidPhoneError{Phone: cfg.Phone}
	}

	db, err := database.Open(cfg.WorkDir)
	if err != nil {
		return nil, err
	}

	devID, err := db.DeviceID()
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	token := cfg.Token
	if token == "" {
		if t, err := db.AuthToken(); err == nil {
			token = t
		}
	}

	clientLogger := cfg.Logger
	if clientLogger == nil {
		clientLogger = logger.Default()
	}

	httpClient := &http.Client{
		Timeout: 5 * time.Minute,
	}

	return &MaxClient{
		cfg:               cfg,
		logger:            clientLogger,
		db:                db,
		httpClient:        httpClient,
		deviceID:          devID,
		token:             token,
		pending:           make(map[int]chan map[string]any),
		incoming:          make(chan map[string]any, 128),
		outgoing:          make(chan map[string]any, 128),
		fileUploadWaiters: make(map[int64]chan map[string]any),
	}, nil
}

// Подключает клиента к WebSocket API Max, выполняет авторизацию и запускает фоновые циклы обработки.
func (c *MaxClient) Start(ctx context.Context) error {
	c.logger.Info("Starting MaxClient", "uri", c.cfg.URI, "phone", c.cfg.Phone)
	ctx, cancel := context.WithCancel(ctx)
	c.bgCancel = cancel

	if err := c.dialWebSocket(ctx); err != nil {
		c.logger.Error("Failed to dial WebSocket", "err", err)
		return err
	}

	c.bgWG.Add(2)
	go func() {
		defer c.bgWG.Done()
		c.recvLoop(ctx)
	}()
	go func() {
		defer c.bgWG.Done()
		c.sendLoop(ctx)
	}()

	if err := c.sessionInit(ctx); err != nil {
		c.logger.Error("SESSION_INIT failed", "err", err)
		cancel()
		c.connMu.Lock()
		if c.ws != nil {
			_ = c.ws.Close()
		}
		c.connMu.Unlock()
		c.bgWG.Wait()
		return err
	}

	select {
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}

	if c.token == "" {
		if c.cfg.Registration {
			c.logger.Info("Starting registration flow")
			if err := c.Register(ctx, c.cfg.FirstName, c.cfg.LastName); err != nil {
				c.logger.Error("Registration failed", "err", err)
				cancel()
				c.connMu.Lock()
				if c.ws != nil {
					_ = c.ws.Close()
				}
				c.connMu.Unlock()
				c.bgWG.Wait()
				return err
			}
		} else {
			c.logger.Info("Starting login flow")
			if err := c.Login(ctx); err != nil {
				c.logger.Error("Login failed", "err", err)
				cancel()
				c.connMu.Lock()
				if c.ws != nil {
					_ = c.ws.Close()
				}
				c.connMu.Unlock()
				c.bgWG.Wait()
				return err
			}
		}
	}

	if err := c.sync(ctx); err != nil {
		if maxErr, ok := err.(*Error); ok && maxErr.Code == "login.token" {
			c.logger.Info("Token invalid, performing re-login")
			c.token = ""
			if c.cfg.Registration {
				c.logger.Info("Starting registration flow")
				if err := c.Register(ctx, c.cfg.FirstName, c.cfg.LastName); err != nil {
					c.logger.Error("Registration failed", "err", err)
					return err
				}
			} else {
				c.logger.Info("Starting login flow")
				if err := c.Login(ctx); err != nil {
					c.logger.Error("Login failed", "err", err)
					return err
				}
			}
			if err := c.sync(ctx); err != nil {
				c.logger.Error("SYNC failed after re-login", "err", err)
				cancel()
				c.connMu.Lock()
				if c.ws != nil {
					_ = c.ws.Close()
				}
				c.connMu.Unlock()
				c.bgWG.Wait()
				return err
			}
		} else {
			c.logger.Error("SYNC failed", "err", err)
			cancel()
			c.connMu.Lock()
			if c.ws != nil {
				_ = c.ws.Close()
			}
			c.connMu.Unlock()
			c.bgWG.Wait()
			return err
		}
	}

	c.logger.Info("Client started successfully")
	for _, h := range c.onStartHandlers {
		h(ctx)
	}

	return nil
}

// Корректно завершает работу MaxClient: останавливает фоновые goroutines,
// закрывает WebSocket-соединение и базу данных сессии.
func (c *MaxClient) Close() error {
	c.logger.Info("Closing MaxClient")
	if c.bgCancel != nil {
		c.bgCancel()
	}
	c.connMu.Lock()
	if c.ws != nil {
		_ = c.ws.Close()
	}
	c.ws = nil
	c.isConnected = false
	c.connMu.Unlock()
	c.bgWG.Wait()

	c.pendingMu.Lock()
	for seq, ch := range c.pending {
		delete(c.pending, seq)
		close(ch)
	}
	c.pendingMu.Unlock()

	c.fileUploadWaitersMu.Lock()
	for fileID, ch := range c.fileUploadWaiters {
		delete(c.fileUploadWaiters, fileID)
		close(ch)
	}
	c.fileUploadWaitersMu.Unlock()

	if c.db != nil {
		if err := c.db.Close(); err != nil {
			c.logger.Error("Failed to close database", "err", err)
			return err
		}
	}
	c.logger.Info("MaxClient closed")
	return nil
}

func (c *MaxClient) nextSeq() int {
	c.seqMu.Lock()
	defer c.seqMu.Unlock()
	c.seq++
	return c.seq
}

// Устанавливает WebSocket‑соединение (только dial, без SESSION_INIT).
func (c *MaxClient) dialWebSocket(ctx context.Context) error {
	c.logger.Debug("Dialing WebSocket", "uri", c.cfg.URI)
	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: time.Duration(constants.DefaultTimeout * float64(time.Second)),
	}

	header := make(http.Header)
	header.Set("Origin", constants.WebsocketOrigin)
	header.Set("User-Agent", constants.DefaultUserAgent)

	ws, _, err := dialer.DialContext(ctx, c.cfg.URI, header)
	if err != nil {
		c.logger.Error("WebSocket dial failed", "err", err, "uri", c.cfg.URI)
		return err
	}

	c.connMu.Lock()
	c.ws = ws
	c.isConnected = true
	c.connMu.Unlock()
	c.logger.Info("WebSocket connected")
	return nil
}

// Выполняет SESSION_INIT handshake.
func (c *MaxClient) sessionInit(ctx context.Context) error {
	c.logger.Debug("Sending SESSION_INIT")
	if err := c.sendAndWait(ctx, enums.OpcodeSessionInit, map[string]any{
		"deviceId": c.deviceID.String(),
		"userAgent": map[string]any{
			"appVersion": constants.DefaultAppVersion,
			"deviceType": constants.DefaultDeviceType,
		},
	}); err != nil {
		c.logger.Error("SESSION_INIT failed", "err", err)
		return err
	}
	c.logger.Debug("SESSION_INIT completed")
	return nil
}

// Выполняет первичный SYNC профиля и чатов.
// Использует OpcodeLogin (19), как в PyMax, а не OpcodeSync (21).
// Возвращает ошибку, если токен невалидный (например, login.token), чтобы можно было выполнить повторный логин.
func (c *MaxClient) sync(ctx context.Context) error {
	c.logger.Debug("Sending SYNC")
	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeLogin, map[string]any{
		"interactive":  true,
		"token":        c.token,
		"chatsSync":    0,
		"contactsSync": 0,
		"presenceSync": 0,
		"draftsSync":   0,
		"chatsCount":   40,
	})
	if err != nil {
		c.logger.Error("SYNC failed", "err", err)
		return err
	}

	if err := HandleError(resp); err != nil {
		c.logger.Warn("SYNC response contains error", "err", err)
		return err
	}

	c.logger.Debug("SYNC completed")
	return nil
}

func (c *MaxClient) reconnect(ctx context.Context) error {
	c.reconnectMu.Lock()
	if c.reconnecting {
		c.reconnectMu.Unlock()
		return fmt.Errorf("reconnection already in progress")
	}
	c.reconnecting = true
	c.reconnectMu.Unlock()

	defer func() {
		c.reconnectMu.Lock()
		c.reconnecting = false
		c.reconnectMu.Unlock()
	}()

	c.logger.Info("Starting reconnection")

	c.connMu.Lock()
	if c.ws != nil {
		_ = c.ws.Close()
		c.ws = nil
	}
	c.isConnected = false
	c.connMu.Unlock()

	if err := c.dialWebSocket(ctx); err != nil {
		c.logger.Error("Failed to dial WebSocket during reconnection", "err", err)
		return fmt.Errorf("dial failed: %w", err)
	}

	if err := c.sessionInit(ctx); err != nil {
		c.logger.Error("SESSION_INIT failed during reconnection", "err", err)
		c.connMu.Lock()
		if c.ws != nil {
			_ = c.ws.Close()
			c.ws = nil
		}
		c.connMu.Unlock()
		return fmt.Errorf("session init failed: %w", err)
	}

	if c.token == "" {
		if c.cfg.Registration {
			c.logger.Info("Performing registration during reconnection")
			if err := c.Register(ctx, c.cfg.FirstName, c.cfg.LastName); err != nil {
				c.logger.Error("Registration failed during reconnection", "err", err)
				c.connMu.Lock()
				if c.ws != nil {
					_ = c.ws.Close()
					c.ws = nil
				}
				c.connMu.Unlock()
				return fmt.Errorf("registration failed: %w", err)
			}
		} else {
			c.logger.Info("Performing login during reconnection")
			if err := c.Login(ctx); err != nil {
				c.logger.Error("Login failed during reconnection", "err", err)
				c.connMu.Lock()
				if c.ws != nil {
					_ = c.ws.Close()
					c.ws = nil
				}
				c.connMu.Unlock()
				return fmt.Errorf("login failed: %w", err)
			}
		}
	}

	if err := c.sync(ctx); err != nil {
		if maxErr, ok := err.(*Error); ok && maxErr.Code == "login.token" {
			c.logger.Info("Token invalid during reconnection, performing re-login")
			c.token = ""
			if c.cfg.Registration {
				c.logger.Info("Starting registration flow during reconnection")
				if err := c.Register(ctx, c.cfg.FirstName, c.cfg.LastName); err != nil {
					c.logger.Error("Registration failed during reconnection", "err", err)
					c.connMu.Lock()
					if c.ws != nil {
						_ = c.ws.Close()
						c.ws = nil
					}
					c.connMu.Unlock()
					return fmt.Errorf("registration failed: %w", err)
				}
			} else {
				c.logger.Info("Starting login flow during reconnection")
				if err := c.Login(ctx); err != nil {
					c.logger.Error("Login failed during reconnection", "err", err)
					c.connMu.Lock()
					if c.ws != nil {
						_ = c.ws.Close()
						c.ws = nil
					}
					c.connMu.Unlock()
					return fmt.Errorf("login failed: %w", err)
				}
			}
			if err := c.sync(ctx); err != nil {
				c.logger.Error("SYNC failed after re-login during reconnection", "err", err)
				c.connMu.Lock()
				if c.ws != nil {
					_ = c.ws.Close()
					c.ws = nil
				}
				c.connMu.Unlock()
				return fmt.Errorf("sync failed: %w", err)
			}
		} else {
			c.logger.Error("SYNC failed during reconnection", "err", err)
			c.connMu.Lock()
			if c.ws != nil {
				_ = c.ws.Close()
				c.ws = nil
			}
			c.connMu.Unlock()
			return fmt.Errorf("sync failed: %w", err)
		}
	}

	c.logger.Info("Reconnection completed successfully")
	return nil
}

// Отправляет запрос с заданным opcode и payload
// и ожидает подтверждающий ответ по тому же seq без разбора содержимого.
func (c *MaxClient) sendAndWait(ctx context.Context, opcode enums.Opcode, payload map[string]any) error {
	seq := c.nextSeq()
	respCh := make(chan map[string]any, 1)

	c.pendingMu.Lock()
	c.pending[seq] = respCh
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, seq)
		c.pendingMu.Unlock()
	}()

	msg := map[string]any{
		"ver":     constants.ProtocolVersion,
		"cmd":     constants.ProtocolCommand,
		"seq":     seq,
		"opcode":  int(opcode),
		"payload": payload,
	}

	select {
	case c.outgoing <- msg:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case <-respCh:

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Непрерывно читает сообщения из WebSocket, разворачивает JSON
// и маршрутизирует ответы либо в ожидающие каналы, либо в обработчики уведомлений.
func (c *MaxClient) recvLoop(ctx context.Context) {
	for {
		c.connMu.RLock()
		ws := c.ws
		c.connMu.RUnlock()
		if ws == nil {
			return
		}

		_, data, err := ws.ReadMessage()
		if err != nil {
			errStr := err.Error()
			isNormalClose := websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) ||
				strings.Contains(errStr, "use of closed network connection") ||
				strings.Contains(errStr, "connection reset by peer")

			c.connMu.Lock()
			c.isConnected = false
			c.connMu.Unlock()

			if isNormalClose {
				c.logger.Debug("WebSocket connection closed", "err", err)
			} else {
				c.logger.Warn("WebSocket read error", "err", err)
			}

			if !c.cfg.Reconnect {
				return
			}

			c.logger.Info("Attempting to reconnect", "delay", c.cfg.ReconnectDelay)
			for {
				select {
				case <-ctx.Done():
					c.logger.Debug("Reconnection cancelled due to context cancellation")
					return
				case <-time.After(c.cfg.ReconnectDelay):
				}

				if err := c.reconnect(ctx); err != nil {
					c.logger.Warn("Reconnection attempt failed", "err", err, "retrying in", c.cfg.ReconnectDelay)
					continue
				}

				c.logger.Info("Reconnection successful, resuming recvLoop")
				c.connMu.RLock()
				ws = c.ws
				c.connMu.RUnlock()
				if ws == nil {
					return
				}
				break
			}
			continue
		}
		var msg map[string]any
		if err := utils.JSONUnmarshal(data, &msg); err != nil {
			c.logger.Warn("Failed to unmarshal message", "err", err)
			continue
		}

		opcodeVal, _ := msg["opcode"].(float64)
		opcode := enums.Opcode(opcodeVal)

		if opcode == enums.OpcodeSync || opcode == enums.OpcodeLogin {
			c.handleSyncResponse(msg)
		}

		if seqVal, ok := msg["seq"].(float64); ok {
			seq := int(seqVal)
			c.pendingMu.Lock()
			ch, ok := c.pending[seq]
			if ok {
				delete(c.pending, seq)
			}
			c.pendingMu.Unlock()
			if ok {

				select {
				case ch <- msg:
				default:
					c.logger.Warn("Pending channel full or receiver not waiting", "seq", seq)
				}
				continue
			}
		}

		if opcode == enums.OpcodeNotifAttach {
			c.handleFileUploadNotification(msg)
		}

		if opcode == enums.OpcodeNotifMessage {
			c.handleMessageNotification(ctx, msg)
		}

		if opcode == enums.OpcodeNotifMsgReactionsChanged {
			c.handleReactionChange(ctx, msg)
		}

		if opcode == enums.OpcodeNotifChat {
			c.handleChatUpdate(ctx, msg)
		}

		select {
		case c.incoming <- msg:
		case <-ctx.Done():
			return
		}
	}
}

// Завершает ожидание загрузки файла или видео по полученному NOTIF_ATTACH.
func (c *MaxClient) handleFileUploadNotification(msg map[string]any) {
	payload, _ := msg["payload"].(map[string]any)
	fileID, _ := payload["fileId"].(float64)
	videoID, _ := payload["videoId"].(float64)

	c.fileUploadWaitersMu.Lock()
	defer c.fileUploadWaitersMu.Unlock()

	if fileID > 0 {
		ch, ok := c.fileUploadWaiters[int64(fileID)]
		if ok {
			delete(c.fileUploadWaiters, int64(fileID))

			select {
			case ch <- msg:
			default:

			}
		}
	}
	if videoID > 0 {
		ch, ok := c.fileUploadWaiters[int64(videoID)]
		if ok {
			delete(c.fileUploadWaiters, int64(videoID))

			select {
			case ch <- msg:
			default:

			}
		}
	}
}

// Обрабатывает NOTIF_MESSAGE, обновляет сообщения
// и вызывает зарегистрированные обработчики новых, отредактированных и удалённых сообщений.
func (c *MaxClient) handleMessageNotification(ctx context.Context, msg map[string]any) {
	payload, _ := msg["payload"].(map[string]any)
	message := &types.Message{}
	if err := utils.FromMap(payload, message); err != nil {
		return
	}

	if message.Status != nil {
		if *message.Status == enums.MessageStatusEdited {
			for _, h := range c.onMessageEditHandlers {
				if h.filter == nil || h.filter.Match(message) {
					go h.handler(ctx, message)
				}
			}
		} else if *message.Status == enums.MessageStatusRemoved {
			for _, h := range c.onMessageDeleteHandlers {
				if h.filter == nil || h.filter.Match(message) {
					go h.handler(ctx, message)
				}
			}
		}
	}

	for _, h := range c.onMessageHandlers {
		if h.filter == nil || h.filter.Match(message) {
			go h.handler(ctx, message)
		}
	}
}

// Обрабатывает NOTIF_MSG_REACTIONS_CHANGED
// и собирает агрегированную информацию о реакциях к сообщению.
func (c *MaxClient) handleReactionChange(ctx context.Context, msg map[string]any) {
	payload, _ := msg["payload"].(map[string]any)
	chatID, _ := payload["chatId"].(float64)
	messageID, _ := payload["messageId"].(string)

	totalCount, _ := payload["totalCount"].(float64)
	yourReaction, _ := payload["yourReaction"].(string)
	countersData, _ := payload["counters"].([]interface{})

	counters := make([]types.ReactionCounter, 0, len(countersData))
	for _, cData := range countersData {
		cMap, _ := cData.(map[string]any)
		counter := types.ReactionCounter{}
		if err := utils.FromMap(cMap, &counter); err == nil {
			counters = append(counters, counter)
		}
	}

	reactionInfo := &types.ReactionInfo{
		TotalCount:   int(totalCount),
		YourReaction: &yourReaction,
		Counters:     counters,
	}

	for _, handler := range c.onReactionChange {
		go handler(ctx, messageID, int64(chatID), reactionInfo)
	}
}

// Обрабатывает NOTIF_CHAT, обновляет кэш чатов
// и вызывает соответствующие обработчики.
func (c *MaxClient) handleChatUpdate(ctx context.Context, msg map[string]any) {
	payload, _ := msg["payload"].(map[string]any)
	chatData, _ := payload["chat"].(map[string]any)
	chat := &types.Chat{}
	if err := utils.FromMap(chatData, chat); err != nil {
		return
	}

	c.updateChatCache(chat)

	for _, handler := range c.onChatUpdate {
		go handler(ctx, chat)
	}
}

// Обрабатывает ответы SYNC/LOGIN и обновляет кэш чатов, диалогов и текущий профиль.
func (c *MaxClient) handleSyncResponse(msg map[string]any) {
	c.logger.Debug("Processing SYNC response", "opcode", msg["opcode"], "seq", msg["seq"])
	payload, _ := msg["payload"].(map[string]any)

	c.stateMu.Lock()
	defer c.stateMu.Unlock()

	seqVal, hasSeq := msg["seq"].(float64)
	if hasSeq && seqVal > 0 {
		c.Chats = make([]types.Chat, 0)
		c.Dialogs = make([]types.Dialog, 0)
		c.Channels = make([]types.Channel, 0)
		c.logger.Debug("Cleared cache before SYNC response processing")
	}

	chatsData, _ := payload["chats"].([]interface{})
	hasError := payload["error"] != nil
	c.logger.Debug("SYNC response received",
		"chatsCount", len(chatsData),
		"hasProfile", payload["profile"] != nil,
		"hasError", hasError)
	if hasError {
		c.logger.Warn("SYNC response contains error", "error", payload["error"])
	}
	for _, chatData := range chatsData {
		chatMap, _ := chatData.(map[string]any)
		chatType, _ := chatMap["type"].(string)

		chat := &types.Chat{}
		if err := utils.FromMap(chatMap, chat); err != nil {
			c.logger.Warn("Failed to parse chat from SYNC", "err", err)
			continue
		}

		switch enums.ChatType(chatType) {
		case enums.ChatTypeDialog:
			dialog := &types.Dialog{}
			if err := utils.FromMap(chatMap, dialog); err == nil {
				c.Dialogs = append(c.Dialogs, *dialog)
				c.logger.Debug("Added dialog from SYNC", "id", dialog.ID)
			}
		case enums.ChatTypeChat:
			c.Chats = append(c.Chats, *chat)
			c.logger.Debug("Added chat from SYNC", "id", chat.ID)
		case enums.ChatTypeChannel:
			c.Channels = append(c.Channels, types.Channel(*chat))
			c.logger.Debug("Added channel from SYNC", "id", chat.ID)
		}
	}

	profile, _ := payload["profile"].(map[string]any)
	contact, _ := profile["contact"].(map[string]any)
	if contact != nil {
		me := &types.Me{}
		if err := utils.FromMap(contact, me); err == nil {
			c.Me = me
			c.logger.Debug("Updated profile from SYNC", "phone", me.Phone)
		}
	}

	c.logger.Debug("SYNC processing completed", "totalChats", len(c.Chats), "totalDialogs", len(c.Dialogs), "totalChannels", len(c.Channels))
}

// Читает сообщения из очереди outgoing и отправляет их в WebSocket до отмены контекста.
func (c *MaxClient) sendLoop(ctx context.Context) {
	c.logger.Debug("Send loop started")
	defer c.logger.Debug("Send loop stopped")
	for {
		select {
		case msg := <-c.outgoing:
			data, err := utils.JSONMarshal(msg)
			if err != nil {
				c.logger.Warn("Failed to marshal message", "err", err)
				continue
			}
			c.connMu.RLock()
			ws := c.ws
			c.connMu.RUnlock()
			if ws == nil {
				c.logger.Debug("WebSocket is nil, stopping send loop")
				return
			}
			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				c.logger.Warn("Failed to write message", "err", err)
				errStr := err.Error()
				isConnectionError := strings.Contains(errStr, "use of closed network connection") ||
					strings.Contains(errStr, "connection reset by peer") ||
					strings.Contains(errStr, "broken pipe") ||
					websocket.IsCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway)
				if isConnectionError {
					c.connMu.Lock()
					c.isConnected = false
					c.connMu.Unlock()
					c.logger.Warn("Connection lost detected in sendLoop", "err", err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// Регистрирует обработчик, который будет вызван после успешного старта клиента и первичного SYNC.
func (c *MaxClient) OnStart(handler func(context.Context)) {
	c.onStartHandlers = append(c.onStartHandlers, handler)
}

// Регистрирует обработчик входящих сообщений с необязательным фильтром по содержимому.
func (c *MaxClient) OnMessage(handler func(context.Context, *types.Message), filter *filters.Filter) {
	c.onMessageHandlers = append(c.onMessageHandlers, messageHandler{
		handler: handler,
		filter:  filter,
	})
}

// Profile возвращает копию текущего профиля пользователя. Потокобезопасен.
func (c *MaxClient) Profile() *types.Me {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.Me
}

// ChatList возвращает копию списка чатов. Потокобезопасен.
func (c *MaxClient) ChatList() []types.Chat {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	result := make([]types.Chat, len(c.Chats))
	copy(result, c.Chats)
	return result
}

// DialogList возвращает копию списка диалогов. Потокобезопасен.
func (c *MaxClient) DialogList() []types.Dialog {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	result := make([]types.Dialog, len(c.Dialogs))
	copy(result, c.Dialogs)
	return result
}

// ChannelList возвращает копию списка каналов. Потокобезопасен.
func (c *MaxClient) ChannelList() []types.Channel {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	result := make([]types.Channel, len(c.Channels))
	copy(result, c.Channels)
	return result
}
