package gomax

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fresh-milkshake/gomax/enums"
	"github.com/fresh-milkshake/gomax/files"
	"github.com/fresh-milkshake/gomax/filters"
	"github.com/fresh-milkshake/gomax/internal/constants"
	"github.com/fresh-milkshake/gomax/internal/payloads"
	"github.com/fresh-milkshake/gomax/internal/utils"
	"github.com/fresh-milkshake/gomax/types"
)

// Инициирует авторизацию по номеру телефона:
// отправляет запрос на получение кода подтверждения и возвращает временный токен.
func (c *MaxClient) RequestCode(ctx context.Context, phone string, language string) (string, error) {
	pl := payloads.RequestCodePayload{
		Phone:    phone,
		Type:     enums.AuthTypeStartAuth,
		Language: language,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return "", err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeAuthRequest, payloadMap)
	if err != nil {
		return "", err
	}

	if err := HandleError(resp); err != nil {
		return "", err
	}

	payload, _ := resp["payload"].(map[string]any)
	token, _ := payload["token"].(string)
	if token == "" {
		return "", fmt.Errorf("token not received")
	}
	return token, nil
}

// Выполняет полный процесс входа: запрашивает код, получает его от пользователя
// и подтверждает авторизацию, сохраняя выданный токен в локальное хранилище.
func (c *MaxClient) Login(ctx context.Context) error {
	tempToken, err := c.RequestCode(ctx, c.cfg.Phone, "ru")
	if err != nil {
		return err
	}

	code, err := c.provideCode(ctx)
	if err != nil {
		return err
	}

	return c.SendCode(ctx, code, tempToken)
}

// Получает код подтверждения из пользовательского колбэка или stdin, если колбэк не задан.
func (c *MaxClient) provideCode(ctx context.Context) (string, error) {
	if c.cfg.CodeProvider != nil {
		return c.cfg.CodeProvider(ctx)
	}

	fmt.Print("Enter verification code: ")
	os.Stdout.Sync()

	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		c.logger.Error("Failed to read verification code", "err", err)
		return "", fmt.Errorf("failed to read verification code: %w", err)
	}

	code := strings.TrimSpace(text)
	if len(code) == 0 {
		c.logger.Error("Empty verification code provided")
		return "", fmt.Errorf("empty verification code")
	}

	if len(code) != 6 {
		c.logger.Warn("Code length is not 6 digits", "length", len(code))
	}

	c.logger.Debug("Verification code received", "length", len(code))
	return code, nil
}

// Подтверждает код верификации, обновляет auth‑токен клиента
// и сохраняет его в базе сессии.
func (c *MaxClient) SendCode(ctx context.Context, code string, token string) error {
	pl := payloads.SendCodePayload{
		Token:         token,
		VerifyCode:    code,
		AuthTokenType: enums.AuthTypeCheckCode,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeAuth, payloadMap)
	if err != nil {
		return err
	}

	if err := HandleError(resp); err != nil {
		return err
	}

	payload, _ := resp["payload"].(map[string]any)
	tokenAttrs, _ := payload["tokenAttrs"].(map[string]any)
	login, _ := tokenAttrs[constants.TokenTypeLogin].(map[string]any)
	authToken, _ := login["token"].(string)
	if authToken == "" {
		return fmt.Errorf("login token not received")
	}

	c.token = authToken
	if err := c.db.UpdateToken(c.deviceID, authToken); err != nil {
		return err
	}
	return nil
}

// Регистрирует нового пользователя по номеру телефона и имени
// и сохраняет выданный регистрацией токен в локальное хранилище.
func (c *MaxClient) Register(ctx context.Context, firstName string, lastName *string) error {
	tempToken, err := c.RequestCode(ctx, c.cfg.Phone, "ru")
	if err != nil {
		return err
	}

	code, err := c.provideCode(ctx)
	if err != nil {
		return err
	}

	plSend := payloads.SendCodePayload{
		Token:         tempToken,
		VerifyCode:    code,
		AuthTokenType: enums.AuthTypeCheckCode,
	}
	sendMap, err := utils.ToMap(plSend)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeAuth, sendMap)
	if err != nil {
		return err
	}
	if err := HandleError(resp); err != nil {
		return err
	}

	sendPayload, _ := resp["payload"].(map[string]any)
	tokenAttrs, _ := sendPayload["tokenAttrs"].(map[string]any)
	regData, _ := tokenAttrs[constants.TokenTypeRegister].(map[string]any)
	registerToken, _ := regData["token"].(string)
	if registerToken == "" {
		return fmt.Errorf("registration token not received")
	}

	pl := payloads.RegisterPayload{
		FirstName: firstName,
		LastName:  lastName,
		Token:     registerToken,
		TokenType: enums.AuthTypeRegister,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err = c.sendAndWaitResponse(ctx, enums.OpcodeAuthConfirm, payloadMap)
	if err != nil {
		return err
	}

	if err = HandleError(resp); err != nil {
		return err
	}

	payload, _ := resp["payload"].(map[string]any)
	authToken, _ := payload["token"].(string)
	if authToken == "" {
		return fmt.Errorf("registration token not received")
	}

	c.token = authToken
	return c.db.UpdateToken(c.deviceID, authToken)
}

// Отправляет текстовое сообщение в указанный чат с поддержкой markdown‑форматирования,
// вложений (фото/файлы/видео) и ответов на существующие сообщения.
func (c *MaxClient) SendMessage(ctx context.Context, text string, chatID int64, notify bool, attachment files.BaseFile, attachments []files.BaseFile, replyTo *int64) (*types.Message, error) {
	var attaches []interface{}

	if attachment != nil && len(attachments) > 0 {
		attachment = nil
	}

	if attachment != nil {
		att, err := c.uploadAttachment(ctx, attachment)
		if err != nil {
			return nil, err
		}
		attaches = append(attaches, att)
	} else if len(attachments) > 0 {
		for _, att := range attachments {
			uploaded, err := c.uploadAttachment(ctx, att)
			if err != nil {
				return nil, err
			}
			attaches = append(attaches, uploaded)
		}
	}

	elements, cleanText := utils.GetElementsFromMarkdown(text)
	if cleanText == "" {
		cleanText = text
	}

	msgElements := make([]payloads.MessageElement, len(elements))
	for i, el := range elements {
		from := 0
		if el.From != nil {
			from = *el.From
		}
		msgElements[i] = payloads.MessageElement{
			Type:   string(el.Type),
			From:   from,
			Length: el.Length,
		}
	}

	var replyLink *payloads.ReplyLink
	if replyTo != nil {
		replyLink = &payloads.ReplyLink{
			Type:      constants.LinkTypeReply,
			MessageID: fmt.Sprintf("%d", *replyTo),
		}
	}

	pl := payloads.SendMessagePayload{
		ChatID: chatID,
		Message: payloads.SendMessagePayloadMessage{
			Text:     cleanText,
			CID:      time.Now().UnixMilli(),
			Elements: msgElements,
			Attaches: attaches,
			Link:     replyLink,
		},
		Notify: notify,
	}

	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeMsgSend, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	msg := &types.Message{}
	if err := utils.FromMap(payload, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// Загружает вложение (фото, файл или видео) и возвращает подходящий payload для отправки сообщения.
func (c *MaxClient) uploadAttachment(ctx context.Context, file files.BaseFile) (interface{}, error) {
	switch f := file.(type) {
	case *files.Photo:
		return c.uploadPhoto(ctx, f)
	case *files.File:
		return c.uploadFile(ctx, f)
	case *files.Video:
		return c.uploadVideo(ctx, f)
	default:
		return nil, fmt.Errorf("unsupported file type")
	}
}

// Выполняет HTTP запрос с retry логикой и экспоненциальным backoff.
// Повторяет запрос только для временных сетевых ошибок.
// createRequest - функция для создания нового Request при каждой попытке.
func (c *MaxClient) doHTTPRequestWithRetry(ctx context.Context, createRequest func() (*http.Request, error)) (*http.Response, error) {
	maxRetries := c.cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = constants.DefaultMaxRetries
	}

	initialDelay := c.cfg.RetryInitialDelay
	if initialDelay == 0 {
		initialDelay = time.Duration(constants.DefaultRetryInitialDelay * float64(time.Second))
	}

	maxDelay := c.cfg.RetryMaxDelay
	if maxDelay == 0 {
		maxDelay = time.Duration(constants.DefaultRetryMaxDelay * float64(time.Second))
	}

	var lastErr error
	delay := initialDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			if delay < maxDelay {
				delay = time.Duration(float64(delay) * 2.0)
				if delay > maxDelay {
					delay = maxDelay
				}
			}

			c.logger.Debug("Retrying HTTP request", "attempt", attempt, "delay", delay)
		}

		req, err := createRequest()
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err == nil {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return resp, nil
			}
			resp.Body.Close()
			if resp.StatusCode >= 500 {
				lastErr = &TemporaryError{Err: fmt.Errorf("server error: %d", resp.StatusCode)}
			} else {
				return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
			}
		} else {
			lastErr = &NetworkError{Err: err}
			if !IsTemporaryError(err) {
				return nil, lastErr
			}
		}

		if attempt < maxRetries {
			c.logger.Warn("HTTP request failed, will retry", "attempt", attempt+1, "err", lastErr)
		}
	}

	return nil, fmt.Errorf("HTTP request failed after %d attempts: %w", maxRetries+1, lastErr)
}

// Резервирует слот загрузки фото, выполняет HTTP‑отправку файла
// и возвращает AttachPhotoPayload с токеном загруженного изображения.
func (c *MaxClient) uploadPhoto(ctx context.Context, photo *files.Photo) (interface{}, error) {
	pl := payloads.UploadPayload{Count: 1}
	payloadMap, _ := utils.ToMap(pl)

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodePhotoUpload, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	url, _ := payload["url"].(string)
	if url == "" {
		return nil, fmt.Errorf("upload URL not received")
	}

	ext, _, err := photo.ValidatePhoto()
	if err != nil {
		return nil, err
	}

	data, err := photo.Read()
	if err != nil {
		return nil, err
	}

	bodyStr := ""
	contentType := ""
	{
		body := &strings.Builder{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", fmt.Sprintf("image.%s", ext))
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(data); err != nil {
			return nil, err
		}
		contentType = writer.FormDataContentType()
		writer.Close()
		bodyStr = body.String()
	}

	httpResp, err := c.doHTTPRequestWithRetry(ctx, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(bodyStr))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", contentType)
		return req, nil
	})
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed with status %d", httpResp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, err
	}

	photos, _ := result["photos"].(map[string]any)
	if len(photos) == 0 {
		return nil, fmt.Errorf("no photos in response")
	}

	var photoData map[string]any
	for _, v := range photos {
		photoData, _ = v.(map[string]any)
		break
	}

	token, _ := photoData["token"].(string)
	if token == "" {
		return nil, fmt.Errorf("photo token not received")
	}

	return payloads.AttachPhotoPayload{
		Type:       enums.AttachTypePhoto,
		PhotoToken: token,
	}, nil
}

// Резервирует слот загрузки файла, отправляет его по выданному URL
// и ожидает подтверждения обработки через NOTIF_ATTACH, возвращая AttachFilePayload.
func (c *MaxClient) uploadFile(ctx context.Context, file *files.File) (interface{}, error) {
	pl := payloads.UploadPayload{Count: 1}
	payloadMap, _ := utils.ToMap(pl)

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeFileUpload, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	info, _ := payload["info"].([]interface{})
	if len(info) == 0 {
		return nil, fmt.Errorf("upload info not received")
	}

	infoMap, _ := info[0].(map[string]any)
	url, _ := infoMap["url"].(string)
	fileID, _ := infoMap["fileId"].(float64)
	if url == "" || fileID == 0 {
		return nil, fmt.Errorf("upload URL or file ID not received")
	}

	data, err := file.Read()
	if err != nil {
		return nil, err
	}

	dataStr := string(data)
	fileName := file.FileName()
	contentRange := fmt.Sprintf("0-%d/%d", len(data)-1, len(data))

	httpResp, err := c.doHTTPRequestWithRetry(ctx, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(dataStr))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		req.Header.Set("Content-Range", contentRange)
		return req, nil
	})
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed with status %d", httpResp.StatusCode)
	}

	waitCh := make(chan map[string]any, 1)
	c.fileUploadWaitersMu.Lock()
	c.fileUploadWaiters[int64(fileID)] = waitCh
	c.fileUploadWaitersMu.Unlock()

	select {
	case <-waitCh:
	case <-ctx.Done():
		c.fileUploadWaitersMu.Lock()
		if ch, ok := c.fileUploadWaiters[int64(fileID)]; ok {
			delete(c.fileUploadWaiters, int64(fileID))
			close(ch)
		}
		c.fileUploadWaitersMu.Unlock()
		return nil, ctx.Err()
	case <-time.After(time.Duration(constants.DefaultTimeout * float64(time.Second))):
		c.fileUploadWaitersMu.Lock()
		if ch, ok := c.fileUploadWaiters[int64(fileID)]; ok {
			delete(c.fileUploadWaiters, int64(fileID))
			close(ch)
		}
		c.fileUploadWaitersMu.Unlock()
		return nil, fmt.Errorf("timeout waiting for file processing")
	}

	return payloads.AttachFilePayload{
		Type:   enums.AttachTypeFile,
		FileID: int64(fileID),
	}, nil
}

// Резервирует слот загрузки видео, отправляет бинарные данные
// и ожидает подтверждения обработки через NOTIF_ATTACH, возвращая VideoAttachPayload.
func (c *MaxClient) uploadVideo(ctx context.Context, video *files.Video) (interface{}, error) {
	pl := payloads.UploadPayload{Count: 1}
	payloadMap, _ := utils.ToMap(pl)

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeVideoUpload, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	info, _ := payload["info"].([]interface{})
	if len(info) == 0 {
		return nil, fmt.Errorf("upload info not received")
	}

	infoMap, _ := info[0].(map[string]any)
	url, _ := infoMap["url"].(string)
	videoID, _ := infoMap["videoId"].(float64)
	token, _ := infoMap["token"].(string)
	if url == "" || videoID == 0 || token == "" {
		return nil, fmt.Errorf("upload URL, video ID or token not received")
	}

	data, err := video.Read()
	if err != nil {
		return nil, err
	}

	dataStr := string(data)
	fileName := video.FileName()
	contentRange := fmt.Sprintf("0-%d/%d", len(data)-1, len(data))

	httpResp, err := c.doHTTPRequestWithRetry(ctx, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(dataStr))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		req.Header.Set("Content-Range", contentRange)
		return req, nil
	})
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed with status %d", httpResp.StatusCode)
	}

	waitCh := make(chan map[string]any, 1)
	c.fileUploadWaitersMu.Lock()
	c.fileUploadWaiters[int64(videoID)] = waitCh
	c.fileUploadWaitersMu.Unlock()

	select {
	case <-waitCh:
	case <-ctx.Done():
		c.fileUploadWaitersMu.Lock()
		if ch, ok := c.fileUploadWaiters[int64(videoID)]; ok {
			delete(c.fileUploadWaiters, int64(videoID))
			close(ch)
		}
		c.fileUploadWaitersMu.Unlock()
		return nil, ctx.Err()
	case <-time.After(time.Duration(constants.DefaultTimeout * float64(time.Second))):
		c.fileUploadWaitersMu.Lock()
		if ch, ok := c.fileUploadWaiters[int64(videoID)]; ok {
			delete(c.fileUploadWaiters, int64(videoID))
			close(ch)
		}
		c.fileUploadWaitersMu.Unlock()
		return nil, fmt.Errorf("timeout waiting for video processing")
	}

	return payloads.VideoAttachPayload{
		Type:    enums.AttachTypeVideo,
		VideoID: int64(videoID),
		Token:   token,
	}, nil
}

// Отправляет команду с указанным opcode и payload
// и блокирующе ждёт полного JSON‑ответа от сервера Max.
func (c *MaxClient) sendAndWaitResponse(ctx context.Context, opcode enums.Opcode, payload map[string]any) (map[string]any, error) {
	seq := c.nextSeq()
	respCh := make(chan map[string]any, 1)

	c.pendingMu.Lock()
	c.pending[seq] = respCh
	c.pendingMu.Unlock()

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
		c.pendingMu.Lock()
		if ch, ok := c.pending[seq]; ok {
			delete(c.pending, seq)
			close(ch)
		}
		c.pendingMu.Unlock()
		return nil, ctx.Err()
	}

	select {
	case resp := <-respCh:
		return resp, nil
	case <-ctx.Done():
		c.pendingMu.Lock()
		if ch, ok := c.pending[seq]; ok {
			delete(c.pending, seq)
			close(ch)
		}
		c.pendingMu.Unlock()
		return nil, ctx.Err()
	case <-time.After(time.Duration(constants.DefaultTimeout * float64(time.Second))):
		c.pendingMu.Lock()
		if ch, ok := c.pending[seq]; ok {
			delete(c.pending, seq)
			close(ch)
		}
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// Редактирует ранее отправленное сообщение, изменяя текст, форматирование
// и вложения, и возвращает обновлённый объект Message.
func (c *MaxClient) EditMessage(ctx context.Context, chatID int64, messageID int64, text string, attachment files.BaseFile, attachments []files.BaseFile) (*types.Message, error) {
	var attaches []interface{}

	if attachment != nil && len(attachments) > 0 {
		attachment = nil
	}

	if attachment != nil {
		att, err := c.uploadAttachment(ctx, attachment)
		if err != nil {
			return nil, err
		}
		attaches = append(attaches, att)
	} else if len(attachments) > 0 {
		for _, att := range attachments {
			uploaded, err := c.uploadAttachment(ctx, att)
			if err != nil {
				return nil, err
			}
			attaches = append(attaches, uploaded)
		}
	}

	elements, cleanText := utils.GetElementsFromMarkdown(text)
	if cleanText == "" {
		cleanText = text
	}

	msgElements := make([]payloads.MessageElement, len(elements))
	for i, el := range elements {
		from := 0
		if el.From != nil {
			from = *el.From
		}
		msgElements[i] = payloads.MessageElement{
			Type:   string(el.Type),
			From:   from,
			Length: el.Length,
		}
	}

	pl := payloads.EditMessagePayload{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      cleanText,
		Elements:  msgElements,
		Attaches:  attaches,
	}

	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeMsgEdit, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	msg := &types.Message{}
	if err := utils.FromMap(payload, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// Удаляет одно или несколько сообщений в чате;
// при forMe=true удаление производится только для текущего пользователя.
func (c *MaxClient) DeleteMessage(ctx context.Context, chatID int64, messageIDs []int64, forMe bool) error {
	pl := payloads.DeleteMessagePayload{
		ChatID:     chatID,
		MessageIDs: messageIDs,
		ForMe:      forMe,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeMsgDelete, payloadMap)
	if err != nil {
		return err
	}

	return HandleError(resp)
}

// Загружает сообщения чата в заданном окне относительно отметки времени fromTime.
func (c *MaxClient) FetchHistory(ctx context.Context, chatID int64, fromTime *int64, forward int, backward int) ([]*types.Message, error) {
	if fromTime == nil {
		now := time.Now().UnixMilli()
		fromTime = &now
	}

	pl := payloads.FetchHistoryPayload{
		ChatID:      chatID,
		FromTime:    fromTime,
		Forward:     forward,
		Backward:    backward,
		GetMessages: true,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatHistory, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	messagesData, _ := payload["messages"].([]interface{})

	messages := make([]*types.Message, 0, len(messagesData))
	for _, msgData := range messagesData {
		msgMap, _ := msgData.(map[string]any)
		msg := &types.Message{}
		if err := utils.FromMap(msgMap, msg); err == nil {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

// Закрепляет указанное сообщение в чате и опционально уведомляет участников.
func (c *MaxClient) PinMessage(ctx context.Context, chatID int64, messageID int64, notifyPin bool) error {
	pl := payloads.PinMessagePayload{
		ChatID:       chatID,
		NotifyPin:    notifyPin,
		PinMessageID: messageID,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatUpdate, payloadMap)
	if err != nil {
		return err
	}

	return HandleError(resp)
}

// Запрашивает метаданные видео‑вложения по chatID, messageID и videoID.
func (c *MaxClient) GetVideoById(ctx context.Context, chatID int64, messageID int64, videoID int64) (*types.VideoRequest, error) {
	pl := payloads.GetVideoPayload{
		ChatID:    chatID,
		MessageID: messageID,
		VideoID:   videoID,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeVideoPlay, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	videoReq := &types.VideoRequest{}
	if err := utils.FromMap(payload, videoReq); err != nil {
		return nil, err
	}
	return videoReq, nil
}

// Запрашивает метаданные файла по chatID, messageID и fileID.
func (c *MaxClient) GetFileById(ctx context.Context, chatID int64, messageID int64, fileID int64) (*types.FileRequest, error) {
	pl := payloads.GetFilePayload{
		ChatID:    chatID,
		MessageID: messageID,
		FileID:    fileID,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeFileDownload, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	fileReq := &types.FileRequest{}
	if err := utils.FromMap(payload, fileReq); err != nil {
		return nil, err
	}
	return fileReq, nil
}

// Добавляет реакцию‑эмодзи к сообщению и возвращает агрегированную информацию о реакциях.
func (c *MaxClient) AddReaction(ctx context.Context, chatID int64, messageID string, reaction string) (*types.ReactionInfo, error) {
	pl := payloads.AddReactionPayload{
		ChatID:    chatID,
		MessageID: messageID,
		Reaction: payloads.ReactionInfoPayload{
			ReactionType: constants.ReactionTypeEmoji,
			ID:           reaction,
		},
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeMsgReaction, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	reactionInfoData, _ := payload["reactionInfo"].(map[string]any)
	reactionInfo := &types.ReactionInfo{}
	if err := utils.FromMap(reactionInfoData, reactionInfo); err != nil {
		return nil, err
	}
	return reactionInfo, nil
}

// Получает агрегированные реакции для набора сообщений в чате.
func (c *MaxClient) GetReactions(ctx context.Context, chatID int64, messageIDs []string) (map[string]*types.ReactionInfo, error) {
	pl := payloads.GetReactionsPayload{
		ChatID:     chatID,
		MessageIDs: messageIDs,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeMsgGetReactions, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	messagesReactions, _ := payload["messagesReactions"].(map[string]any)

	result := make(map[string]*types.ReactionInfo)
	for msgID, reactionData := range messagesReactions {
		reactionMap, _ := reactionData.(map[string]any)
		reactionInfo := &types.ReactionInfo{}
		if err := utils.FromMap(reactionMap, reactionInfo); err == nil {
			result[msgID] = reactionInfo
		}
	}

	return result, nil
}

// Удаляет реакцию текущего пользователя с сообщения
// и возвращает обновлённую информацию о реакциях.
func (c *MaxClient) RemoveReaction(ctx context.Context, chatID int64, messageID string) (*types.ReactionInfo, error) {
	pl := payloads.RemoveReactionPayload{
		ChatID:    chatID,
		MessageID: messageID,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeMsgCancelReaction, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	reactionInfoData, _ := payload["reactionInfo"].(map[string]any)
	reactionInfo := &types.ReactionInfo{}
	if err := utils.FromMap(reactionInfoData, reactionInfo); err != nil {
		return nil, err
	}
	return reactionInfo, nil
}

// Получает информацию о пользователе по его ID, используя GetUsers.
func (c *MaxClient) GetUser(ctx context.Context, userID int64) (*types.User, error) {
	users, err := c.GetUsers(ctx, []int64{userID})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return users[0], nil
}

// Получает информацию о нескольких пользователях по их ID через WebSocket API Max.
func (c *MaxClient) GetUsers(ctx context.Context, userIDs []int64) ([]*types.User, error) {
	pl := payloads.FetchContactsPayload{
		ContactIDs: userIDs,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeContactInfo, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	contactsData, _ := payload["contacts"].([]interface{})

	users := make([]*types.User, 0, len(contactsData))
	for _, contactData := range contactsData {
		contactMap, _ := contactData.(map[string]any)
		user := &types.User{}
		if err := utils.FromMap(contactMap, user); err == nil {
			users = append(users, user)
		}
	}

	return users, nil
}

// Ищет пользователя по номеру телефона и возвращает объект User при успехе.
func (c *MaxClient) SearchByPhone(ctx context.Context, phone string) (*types.User, error) {
	pl := payloads.SearchByPhonePayload{
		Phone: phone,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeContactInfoByPhone, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	contactData, _ := payload["contact"].(map[string]any)
	user := &types.User{}
	if err := utils.FromMap(contactData, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Добавляет пользователя в список контактов текущего аккаунта.
func (c *MaxClient) AddContact(ctx context.Context, contactID int64) (*types.Contact, error) {
	pl := payloads.ContactActionPayload{
		ContactID: contactID,
		Action:    enums.ContactActionAdd,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeContactUpdate, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	contactData, _ := payload["contact"].(map[string]any)
	contact := &types.Contact{}
	if err := utils.FromMap(contactData, contact); err != nil {
		return nil, err
	}
	return contact, nil
}

// Удаляет пользователя из списка контактов текущего аккаунта.
func (c *MaxClient) RemoveContact(ctx context.Context, contactID int64) error {
	pl := payloads.ContactActionPayload{
		ContactID: contactID,
		Action:    enums.ContactActionRemove,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeContactUpdate, payloadMap)
	if err != nil {
		return err
	}

	return HandleError(resp)
}

// Вычисляет детерминированный идентификатор диалога между двумя пользователями.
func (c *MaxClient) GetChatId(firstUserID int64, secondUserID int64) int64 {
	return firstUserID ^ secondUserID
}

// Получает список всех активных сессий аккаунта (устройства, платформы и т.п.).
func (c *MaxClient) GetSessions(ctx context.Context) ([]*types.Session, error) {
	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeSessionsInfo, map[string]any{})
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	sessionsData, _ := payload["sessions"].([]interface{})

	sessions := make([]*types.Session, 0, len(sessionsData))
	for _, sessionData := range sessionsData {
		sessionMap, _ := sessionData.(map[string]any)
		session := &types.Session{}
		if err := utils.FromMap(sessionMap, session); err == nil {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// Пытается получить информацию о канале по его публичному имени.
func (c *MaxClient) ResolveChannelByName(ctx context.Context, name string) error {
	pl := payloads.ResolveLinkPayload{
		Link: fmt.Sprintf("https://max.ru/%s", name),
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeLinkInfo, payloadMap)
	if err != nil {
		return err
	}

	return HandleError(resp)
}

// Присоединяется к публичному каналу по полной ссылке-просмотру.
func (c *MaxClient) JoinChannel(ctx context.Context, link string) error {
	pl := payloads.JoinChatPayload{
		Link: link,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatJoin, payloadMap)
	if err != nil {
		return err
	}

	return HandleError(resp)
}

// Загружает участников канала или группы с поддержкой маркера пагинации.
func (c *MaxClient) LoadMembers(ctx context.Context, chatID int64, marker *int, count int) ([]*types.Member, *int, error) {
	if marker == nil {
		zero := 0
		marker = &zero
	}
	if count == 0 {
		count = constants.DefaultChatMembers
	}

	pl := payloads.GetGroupMembersPayload{
		Type:   constants.MemberTypeMember,
		Marker: marker,
		ChatID: chatID,
		Count:  count,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatMembers, payloadMap)
	if err != nil {
		return nil, nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	membersData, _ := payload["members"].([]interface{})
	markerVal, _ := payload["marker"].(float64)

	members := make([]*types.Member, 0, len(membersData))
	for _, memberData := range membersData {
		memberMap, _ := memberData.(map[string]any)
		member := &types.Member{}
		if err := utils.FromMap(memberMap, member); err == nil {
			members = append(members, member)
		}
	}

	var nextMarker *int
	if markerVal > 0 {
		m := int(markerVal)
		nextMarker = &m
	}

	return members, nextMarker, nil
}

// Выполняет поиск участников канала или группы по строке запроса.
func (c *MaxClient) FindMembers(ctx context.Context, chatID int64, query string) ([]*types.Member, *int, error) {
	pl := payloads.SearchGroupMembersPayload{
		Type:   constants.MemberTypeMember,
		Query:  query,
		ChatID: chatID,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatMembers, payloadMap)
	if err != nil {
		return nil, nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	membersData, _ := payload["members"].([]interface{})
	markerVal, _ := payload["marker"].(float64)

	members := make([]*types.Member, 0, len(membersData))
	for _, memberData := range membersData {
		memberMap, _ := memberData.(map[string]any)
		member := &types.Member{}
		if err := utils.FromMap(memberMap, member); err == nil {
			members = append(members, member)
		}
	}

	var nextMarker *int
	if markerVal > 0 {
		m := int(markerVal)
		nextMarker = &m
	}

	return members, nextMarker, nil
}

// Получает информацию о нескольких чатах по их ID
// и обновляет локальный кэш чатов клиента.
func (c *MaxClient) GetChats(ctx context.Context, chatIDs []int64) ([]*types.Chat, error) {
	pl := payloads.GetChatInfoPayload{
		ChatIDs: chatIDs,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatInfo, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	chatsData, _ := payload["chats"].([]interface{})

	chats := make([]*types.Chat, 0, len(chatsData))
	for _, chatData := range chatsData {
		chatMap, _ := chatData.(map[string]any)
		chat := &types.Chat{}
		if err := utils.FromMap(chatMap, chat); err == nil {
			chats = append(chats, chat)
			c.updateChatCache(chat)
		}
	}

	return chats, nil
}

// Получает информацию об одном чате по его ID, используя GetChats.
func (c *MaxClient) GetChat(ctx context.Context, chatID int64) (*types.Chat, error) {
	chats, err := c.GetChats(ctx, []int64{chatID})
	if err != nil {
		return nil, err
	}
	if len(chats) == 0 {
		return nil, fmt.Errorf("chat not found")
	}
	return chats[0], nil
}

// Обновляет или добавляет запись о чате в локальный кэш клиента.
// Метод потокобезопасен.
func (c *MaxClient) updateChatCache(chat *types.Chat) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()

	for i, existing := range c.Chats {
		if existing.ID == chat.ID {
			c.Chats[i] = *chat
			return
		}
	}
	c.Chats = append(c.Chats, *chat)
}

// Создаёт новый групповой чат с указанным списком участников
// и возвращает созданный чат и служебное сообщение о событии создания.
func (c *MaxClient) CreateGroup(ctx context.Context, name string, participantIDs []int64, notify bool) (*types.Chat, *types.Message, error) {
	attach := payloads.CreateGroupAttach{
		Type:     constants.AttachTypeControl,
		Event:    constants.EventNew,
		ChatType: constants.ChatTypeChat,
		Title:    name,
		UserIDs:  participantIDs,
	}

	pl := payloads.CreateGroupPayload{
		Message: payloads.CreateGroupMessage{
			CID:      time.Now().UnixMilli(),
			Attaches: []payloads.CreateGroupAttach{attach},
		},
		Notify: notify,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeMsgSend, payloadMap)
	if err != nil {
		return nil, nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	chatData, _ := payload["chat"].(map[string]any)
	chat := &types.Chat{}
	if err := utils.FromMap(chatData, chat); err != nil {
		return nil, nil, err
	}

	msg := &types.Message{}
	if err := utils.FromMap(payload, msg); err != nil {
		return nil, nil, err
	}

	c.updateChatCache(chat)
	return chat, msg, nil
}

// Приглашает указанных пользователей в группу
// и при необходимости открывает им историю сообщений.
func (c *MaxClient) InviteUsersToGroup(ctx context.Context, chatID int64, userIDs []int64, showHistory bool) error {
	pl := payloads.InviteUsersPayload{
		ChatID:      chatID,
		UserIDs:     userIDs,
		ShowHistory: showHistory,
		Operation:   constants.OperationAdd,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatMembersUpdate, payloadMap)
	if err != nil {
		return err
	}

	if err := HandleError(resp); err != nil {
		return err
	}

	payload, _ := resp["payload"].(map[string]any)
	chatData, _ := payload["chat"].(map[string]any)
	chat := &types.Chat{}
	if err := utils.FromMap(chatData, chat); err == nil {
		c.updateChatCache(chat)
	}

	return nil
}

// Удаляет пользователей из группы и при необходимости очищает их историю сообщений.
func (c *MaxClient) RemoveUsersFromGroup(ctx context.Context, chatID int64, userIDs []int64, cleanMsgPeriod int) error {
	pl := payloads.RemoveUsersPayload{
		ChatID:         chatID,
		UserIDs:        userIDs,
		Operation:      constants.OperationRemove,
		CleanMsgPeriod: cleanMsgPeriod,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatMembersUpdate, payloadMap)
	if err != nil {
		return err
	}

	if err := HandleError(resp); err != nil {
		return err
	}

	payload, _ := resp["payload"].(map[string]any)
	chatData, _ := payload["chat"].(map[string]any)
	chat := &types.Chat{}
	if err := utils.FromMap(chatData, chat); err == nil {
		c.updateChatCache(chat)
	}

	return nil
}

// Изменяет административные настройки группы
// (кто может пинить сообщения, добавлять участников и управлять ссылками).
func (c *MaxClient) ChangeGroupSettings(ctx context.Context, chatID int64, allCanPinMessage *bool, onlyOwnerCanChangeIconTitle *bool, onlyAdminCanAddMember *bool, onlyAdminCanCall *bool, membersCanSeePrivateLink *bool) error {
	pl := payloads.ChangeGroupSettingsPayload{
		ChatID: chatID,
		Options: payloads.ChangeGroupSettingsOptions{
			AllCanPinMessage:            allCanPinMessage,
			OnlyOwnerCanChangeIconTitle: onlyOwnerCanChangeIconTitle,
			OnlyAdminCanAddMember:       onlyAdminCanAddMember,
			OnlyAdminCanCall:            onlyAdminCanCall,
			MembersCanSeePrivateLink:    membersCanSeePrivateLink,
		},
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatUpdate, payloadMap)
	if err != nil {
		return err
	}

	if err := HandleError(resp); err != nil {
		return err
	}

	payload, _ := resp["payload"].(map[string]any)
	chatData, _ := payload["chat"].(map[string]any)
	chat := &types.Chat{}
	if err := utils.FromMap(chatData, chat); err == nil {
		c.updateChatCache(chat)
	}

	return nil
}

// Изменяет основные поля профиля группы: название и описание.
func (c *MaxClient) ChangeGroupProfile(ctx context.Context, chatID int64, name *string, description *string) error {
	pl := payloads.ChangeGroupProfilePayload{
		ChatID:      chatID,
		Theme:       name,
		Description: description,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatUpdate, payloadMap)
	if err != nil {
		return err
	}

	if err := HandleError(resp); err != nil {
		return err
	}

	payload, _ := resp["payload"].(map[string]any)
	chatData, _ := payload["chat"].(map[string]any)
	chat := &types.Chat{}
	if err := utils.FromMap(chatData, chat); err == nil {
		c.updateChatCache(chat)
	}

	return nil
}

// Присоединяется к группе по ссылке приглашения формата .../join/<token>
// и возвращает объект чата после успешного входа.
func (c *MaxClient) JoinGroup(ctx context.Context, link string) (*types.Chat, error) {
	idx := strings.Index(link, "join/")
	if idx == -1 {
		return nil, fmt.Errorf("invalid group link")
	}
	proceedLink := link[idx:]

	pl := payloads.JoinChatPayload{
		Link: proceedLink,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatJoin, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	chatData, _ := payload["chat"].(map[string]any)
	chat := &types.Chat{}
	if err := utils.FromMap(chatData, chat); err != nil {
		return nil, err
	}

	c.updateChatCache(chat)
	return chat, nil
}

// Отзывает текущую приватную ссылку приглашения и создаёт новую для указанной группы.
func (c *MaxClient) ReworkInviteLink(ctx context.Context, chatID int64) (*types.Chat, error) {
	pl := payloads.ReworkInviteLinkPayload{
		RevokePrivateLink: true,
		ChatID:            chatID,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeChatUpdate, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	chatData, _ := payload["chat"].(map[string]any)
	chat := &types.Chat{}
	if err := utils.FromMap(chatData, chat); err != nil {
		return nil, fmt.Errorf("chat data missing in response")
	}

	c.updateChatCache(chat)
	return chat, nil
}

// Изменяет профиль текущего пользователя (имя, фамилию и статус/описание).
func (c *MaxClient) ChangeProfile(ctx context.Context, firstName string, lastName *string, description *string) error {
	pl := payloads.ChangeProfilePayload{
		FirstName:   firstName,
		LastName:    lastName,
		Description: description,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeProfile, payloadMap)
	if err != nil {
		return err
	}

	return HandleError(resp)
}

// Создаёт пользовательскую папку (фильтр) для выбранных чатов
// и возвращает информацию об обновлении списка папок.
func (c *MaxClient) CreateFolder(ctx context.Context, title string, chatInclude []int64, folderFilters []interface{}) (*types.FolderUpdate, error) {
	pl := payloads.CreateFolderPayload{
		ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
		Title:   title,
		Include: chatInclude,
		Filters: folderFilters,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeFoldersUpdate, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	folderUpdate := &types.FolderUpdate{}
	if err := utils.FromMap(payload, folderUpdate); err != nil {
		return nil, err
	}
	return folderUpdate, nil
}

// Возвращает список папок пользователя и текущий маркер синхронизации.
func (c *MaxClient) GetFolders(ctx context.Context, folderSync int) (*types.FolderList, error) {
	pl := payloads.GetFolderPayload{
		FolderSync: folderSync,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeFoldersGet, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	folderList := &types.FolderList{}
	if err := utils.FromMap(payload, folderList); err != nil {
		return nil, err
	}
	return folderList, nil
}

// Изменяет параметры существующей папки (название, список чатов, фильтры и опции).
func (c *MaxClient) UpdateFolder(ctx context.Context, folderID string, title string, chatInclude []int64, folderFilters []interface{}, options []interface{}) (*types.FolderUpdate, error) {
	pl := payloads.UpdateFolderPayload{
		ID:      folderID,
		Title:   title,
		Include: chatInclude,
		Filters: folderFilters,
		Options: options,
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeFoldersUpdate, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	folderUpdate := &types.FolderUpdate{}
	if err := utils.FromMap(payload, folderUpdate); err != nil {
		return nil, err
	}
	return folderUpdate, nil
}

// Удаляет папку по идентификатору и возвращает результат обновления порядка папок.
func (c *MaxClient) DeleteFolder(ctx context.Context, folderID string) (*types.FolderUpdate, error) {
	pl := payloads.DeleteFolderPayload{
		FolderIDs: []string{folderID},
	}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendAndWaitResponse(ctx, enums.OpcodeFoldersDelete, payloadMap)
	if err != nil {
		return nil, err
	}

	if err := HandleError(resp); err != nil {
		return nil, err
	}

	payload, _ := resp["payload"].(map[string]any)
	folderUpdate := &types.FolderUpdate{}
	if err := utils.FromMap(payload, folderUpdate); err != nil {
		return nil, err
	}
	return folderUpdate, nil
}

// Регистрирует обработчик для уведомлений об отредактированных сообщениях с опциональным фильтром.
func (c *MaxClient) OnMessageEdit(handler func(context.Context, *types.Message), filter *filters.Filter) {
	c.onMessageEditHandlers = append(c.onMessageEditHandlers, messageHandler{
		handler: handler,
		filter:  filter,
	})
}

// Регистрирует обработчик для уведомлений об удалённых сообщениях с опциональным фильтром.
func (c *MaxClient) OnMessageDelete(handler func(context.Context, *types.Message), filter *filters.Filter) {
	c.onMessageDeleteHandlers = append(c.onMessageDeleteHandlers, messageHandler{
		handler: handler,
		filter:  filter,
	})
}

// Регистрирует обработчик изменения реакций на сообщения.
func (c *MaxClient) OnReactionChange(handler func(context.Context, string, int64, *types.ReactionInfo)) {
	c.onReactionChange = append(c.onReactionChange, handler)
}

// Регистрирует обработчик уведомлений об обновлении чатов.
func (c *MaxClient) OnChatUpdate(handler func(context.Context, *types.Chat)) {
	c.onChatUpdate = append(c.onChatUpdate, handler)
}
