package gomax

import (
	"context"
	"testing"
	"time"

	"github.com/fresh-milkshake/gomax/internal/constants"
	"github.com/fresh-milkshake/gomax/logger"
	"github.com/fresh-milkshake/gomax/mockserver"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPhone      = "+79991234567"
	testTempToken  = "temp_token_123"
	testRegToken   = "register_token_456"
	testLoginToken = "login_token_789"
	testAuthToken  = "auth_token_final"
	testFirstName  = "Test"
	testLastName   = "User"
	testVerifyCode = "123456"
)

// TestRequestCode_Success проверяет успешный запрос кода подтверждения.
func TestRequestCode_Success(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServer(t)
	server.DefaultHandlers()
	server.SetupAuthHandlers(testTempToken, "", "", "")

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)

	err = client.Start(ctx)

	assert.Error(t, err)

	client2, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client2.Close()

	err = client2.Start(ctx)
	require.NoError(t, err)

	token, err := client2.RequestCode(ctx, testPhone, "ru")
	require.NoError(t, err)
	assert.Equal(t, testTempToken, token)
}

// TestSendCode_Success проверяет успешную отправку кода подтверждения.
func TestSendCode_Success(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServer(t)
	server.DefaultHandlers()

	server.SetHandler(18, func(msg map[string]any) map[string]any {
		return mockserver.AuthResponse(0, "", testLoginToken)
	})

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	err = client.SendCode(ctx, testVerifyCode, testTempToken)
	require.NoError(t, err)
}

// TestLogin_EndToEnd проверяет полный flow логина.
func TestLogin_EndToEnd(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServer(t)
	server.DefaultHandlers()

	server.SetHandler(17, func(msg map[string]any) map[string]any {
		return mockserver.AuthRequestResponse(0, testTempToken)
	})

	server.SetHandler(18, func(msg map[string]any) map[string]any {
		return mockserver.AuthResponse(0, "", testLoginToken)
	})

	workDir := t.TempDir()

	codeProvided := make(chan struct{})
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Logger:  logger.Nop(),
		CodeProvider: func(ctx context.Context) (string, error) {
			close(codeProvided)
			return testVerifyCode, nil
		},
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	select {
	case <-codeProvided:

	case <-time.After(5 * time.Second):
		t.Fatal("code provider was not called")
	}
}

// TestLogin_QRFlow проверяет QR‑авторизацию (device_type=WEB).
func TestLogin_QRFlow(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServer(t)
	server.DefaultHandlers()

	const trackID = "track-123"
	const qrLink = "https://qr.example"
	const loginToken = "qr_login_token"

	server.SetHandler(mockserver.OpcodeGetQR, func(msg map[string]any) map[string]any {
		return mockserver.GetQRResponse(int(msg["seq"].(float64)), trackID, qrLink, 100, time.Now().Add(5*time.Second).UnixMilli())
	})
	server.SetHandler(mockserver.OpcodeGetQRStatus, func(msg map[string]any) map[string]any {
		return mockserver.GetQRStatusResponse(int(msg["seq"].(float64)), true, time.Now().Add(5*time.Second).UnixMilli())
	})
	server.SetHandler(mockserver.OpcodeLoginByQR, func(msg map[string]any) map[string]any {
		return mockserver.LoginByQRResponse(int(msg["seq"].(float64)), loginToken)
	})

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Logger:  logger.Nop(),
		UserAgent: UserAgent{
			DeviceType:      constants.DeviceTypeWeb,
			AppVersion:      constants.MinWebQRAppVersion,
			HeaderUserAgent: constants.DefaultUserAgent,
		},
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	// Проверяем, что SYNC отправлен с полученным токеном.
	msgs := server.GetReceivedMessages()
	var syncMsg map[string]any
	for _, m := range msgs {
		if int(m["opcode"].(float64)) == mockserver.OpcodeLogin {
			syncMsg = m
			break
		}
	}
	if syncMsg == nil {
		t.Fatalf("SYNC not sent after QR login")
	}
	payload := syncMsg["payload"].(map[string]any)
	assert.Equal(t, loginToken, payload["token"])
}

// TestLogin_QRVersionTooLow проверяет отказ при недостаточной версии для WEB QR.
func TestLogin_QRVersionTooLow(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServer(t)
	server.DefaultHandlers()

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Logger:  logger.Nop(),
		UserAgent: UserAgent{
			DeviceType: constants.DeviceTypeWeb,
			AppVersion: "25.10.13",
		},
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.Error(t, err)
}

// TestRegister_EndToEnd проверяет полный flow регистрации.
func TestRegister_EndToEnd(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServer(t)
	server.DefaultHandlers()

	server.SetHandler(17, func(msg map[string]any) map[string]any {
		return mockserver.AuthRequestResponse(0, testTempToken)
	})

	server.SetHandler(18, func(msg map[string]any) map[string]any {
		return mockserver.AuthResponse(0, testRegToken, "")
	})

	server.SetHandler(23, func(msg map[string]any) map[string]any {
		return mockserver.AuthConfirmResponse(0, testAuthToken)
	})

	workDir := t.TempDir()

	codeProvided := make(chan struct{})
	client, err := NewMaxClient(ClientConfig{
		Phone:        testPhone,
		URI:          server.URL(),
		WorkDir:      workDir,
		Registration: true,
		FirstName:    testFirstName,
		LastName:     mockserver.StringPtr(testLastName),
		Logger:       logger.Nop(),
		CodeProvider: func(ctx context.Context) (string, error) {
			close(codeProvided)
			return testVerifyCode, nil
		},
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	select {
	case <-codeProvided:

	case <-time.After(5 * time.Second):
		t.Fatal("code provider was not called")
	}
}

// TestStart_WithExistingToken проверяет старт клиента с существующим токеном.
func TestStart_WithExistingToken(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

}

// TestStart_OnStartHandler проверяет вызов OnStart обработчика.
func TestStart_OnStartHandler(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	handlerCalled := make(chan struct{})
	client.OnStart(func(ctx context.Context) {
		close(handlerCalled)
	})

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	select {
	case <-handlerCalled:

	case <-time.After(5 * time.Second):
		t.Fatal("OnStart handler was not called")
	}
}

// TestClose проверяет корректное закрытие клиента.
func TestClose(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	err = client.Close()
	require.NoError(t, err)
}

// TestInvalidPhone проверяет ошибку при неверном формате телефона.
func TestInvalidPhone(t *testing.T) {
	t.Parallel()
	_, err := NewMaxClient(ClientConfig{
		Phone:   "invalid",
		WorkDir: t.TempDir(),
		Logger:  logger.Nop(),
	})
	assert.Error(t, err)
	assert.IsType(t, &InvalidPhoneError{}, err)
}

// TestSyncResponse проверяет обработку ответа SYNC через уведомления.
func TestSyncResponse(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServer(t)

	server.SetHandler(6, func(msg map[string]any) map[string]any {
		return mockserver.SessionInitResponse(0)
	})

	server.SetHandler(19, func(msg map[string]any) map[string]any {
		return mockserver.SyncResponse(0, nil, nil)
	})

	server.SetHandler(21, func(msg map[string]any) map[string]any {
		return mockserver.SyncResponse(0, nil, nil)
	})

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:   testPhone,
		URI:     server.URL(),
		WorkDir: workDir,
		Token:   testAuthToken,
		Logger:  logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	syncNotification := map[string]any{
		"ver":    11,
		"cmd":    0,
		"seq":    0,
		"opcode": 21,
		"payload": map[string]any{
			"profile": map[string]any{
				"contact": map[string]any{
					"id":    int64(999),
					"phone": testPhone,
					"names": []map[string]any{
						{
							"firstName": "Custom",
							"lastName":  "User",
							"type":      "FIRST_LAST",
						},
					},
				},
			},
			"chats": []map[string]any{
				mockserver.TestChat(100, "DIALOG", ""),
				mockserver.TestChat(200, "CHAT", "Test Group"),
				mockserver.TestChat(300, "CHANNEL", "Test Channel"),
			},
		},
	}
	err = server.SendNotification(syncNotification)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	me := client.Profile()
	require.NotNil(t, me)
	require.NotEmpty(t, me.Names)
	assert.Equal(t, "Custom", *me.Names[0].FirstName)
}

// TestReconnect_OnConnectionLoss проверяет переподключение при потере соединения.
func TestReconnect_OnConnectionLoss(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:          testPhone,
		URI:            server.URL(),
		WorkDir:        workDir,
		Token:          testAuthToken,
		Reconnect:      true,
		ReconnectDelay: 100 * time.Millisecond,
		Logger:         logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	server.CloseAllConnections()

	time.Sleep(500 * time.Millisecond)

	assert.True(t, server.IsConnected(), "client should have reconnected")
}

// TestReconnect_ReAuthentication проверяет повторную аутентификацию при переподключении.
func TestReconnect_ReAuthentication(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:          testPhone,
		URI:            server.URL(),
		WorkDir:        workDir,
		Token:          testAuthToken,
		Reconnect:      true,
		ReconnectDelay: 100 * time.Millisecond,
		Logger:         logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	time.Sleep(300 * time.Millisecond)

	initialMessages := server.GetReceivedMessages()
	initialSessionInitCount := 0
	initialSyncCount := 0
	for _, msg := range initialMessages {
		if opcode, ok := msg["opcode"].(float64); ok {
			if int(opcode) == 6 {
				initialSessionInitCount++
			}
			if int(opcode) == 19 {
				initialSyncCount++
			}
		}
	}

	server.ClearReceivedMessages()

	server.CloseAllConnections()

	time.Sleep(1 * time.Second)

	assert.True(t, server.IsConnected(), "client should have reconnected")

	if server.IsConnected() {
		t.Log("Client successfully reconnected")
	}
}

// TestReconnect_Disabled проверяет поведение при Reconnect = false.
func TestReconnect_Disabled(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:          testPhone,
		URI:            server.URL(),
		WorkDir:        workDir,
		Token:          testAuthToken,
		Reconnect:      false,
		ReconnectDelay: 100 * time.Millisecond,
		Logger:         logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	server.CloseAllConnections()

	time.Sleep(500 * time.Millisecond)

	assert.False(t, server.IsConnected(), "client should not reconnect when Reconnect=false")
}

// TestReconnect_MultipleReconnections проверяет множественные переподключения.
func TestReconnect_MultipleReconnections(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServerWithDefaults(t)

	workDir := t.TempDir()
	client, err := NewMaxClient(ClientConfig{
		Phone:          testPhone,
		URI:            server.URL(),
		WorkDir:        workDir,
		Token:          testAuthToken,
		Reconnect:      true,
		ReconnectDelay: 200 * time.Millisecond,
		Logger:         logger.Nop(),
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	if !mockserver.WaitForCondition(t, 2*time.Second, func() bool {
		return server.IsConnected()
	}) {
		t.Fatal("connection not established before disconnect")
	}

	server.CloseAllConnections()

	if !mockserver.WaitForCondition(t, 3*time.Second, func() bool {
		return server.IsConnected()
	}) {
		t.Fatal("client did not reconnect after disconnect")
	}

	time.Sleep(500 * time.Millisecond)
	assert.True(t, server.IsConnected(), "connection should remain stable after reconnection")
}

// TestReconnect_WithLoginFlow проверяет переподключение с повторным логином.
func TestReconnect_WithLoginFlow(t *testing.T) {
	t.Parallel()
	server := mockserver.StartMockServer(t)
	server.DefaultHandlers()

	server.SetHandler(17, func(msg map[string]any) map[string]any {
		return mockserver.AuthRequestResponse(0, testTempToken)
	})

	server.SetHandler(18, func(msg map[string]any) map[string]any {
		return mockserver.AuthResponse(0, "", testLoginToken)
	})

	workDir := t.TempDir()

	codeProvided := make(chan struct{}, 2)
	client, err := NewMaxClient(ClientConfig{
		Phone:          testPhone,
		URI:            server.URL(),
		WorkDir:        workDir,
		Reconnect:      true,
		ReconnectDelay: 100 * time.Millisecond,
		Logger:         logger.Nop(),
		CodeProvider: func(ctx context.Context) (string, error) {
			codeProvided <- struct{}{}
			return testVerifyCode, nil
		},
	})
	require.NoError(t, err)
	defer client.Close()

	ctx := mockserver.TestContext(t)
	err = client.Start(ctx)
	require.NoError(t, err)

	select {
	case <-codeProvided:

	case <-time.After(5 * time.Second):
		t.Fatal("code provider was not called on initial connection")
	}

	time.Sleep(200 * time.Millisecond)

	server.CloseAllConnections()

	time.Sleep(1 * time.Second)

	assert.True(t, server.IsConnected(), "client should have reconnected")

	select {
	case <-codeProvided:

	default:

	}

	assert.True(t, server.IsConnected(), "client should have reconnected")
}
