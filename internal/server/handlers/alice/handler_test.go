package alice_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/capcom6/alice-companion/internal/alice"
	"github.com/capcom6/alice-companion/internal/server/auth"
	alicehttp "github.com/capcom6/alice-companion/internal/server/handlers/alice"
	"github.com/go-core-fx/fiberfx"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const testSecret = "test-secret"

func newTestApp(t *testing.T, secret string) *fiber.App {
	t.Helper()

	return newTestAppWithLogger(t, secret, zap.NewNop())
}

func newTestAppWithLogger(t *testing.T, secret string, logger *zap.Logger) *fiber.App {
	t.Helper()

	app := fiber.New(fiber.Config{ErrorHandler: fiberfx.NewJSONErrorHandler(logger)})
	alicehttp.New(alice.Config{SharedSecret: secret}, validator.New(), logger).Register(app.Group("/api/v1"))

	return app
}

func doPost(t *testing.T, app *fiber.App, secret, body string) (int, []byte) {
	t.Helper()

	return doPostPath(t, app, secret, "/api/v1/alice/webhook", body)
}

func doPostPath(t *testing.T, app *fiber.App, secret, path, body string) (int, []byte) {
	t.Helper()

	if secret != "" {
		path += "?" + auth.SecretQuery + "=" + url.QueryEscape(secret)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		path,
		strings.NewReader(body),
	)
	require.NoError(t, err)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	return resp.StatusCode, data
}

func marshalRequest(t *testing.T, req alice.WebhookRequest) string {
	t.Helper()

	data, err := json.Marshal(req)
	require.NoError(t, err)

	return string(data)
}

func decodeResponse(t *testing.T, data []byte) alice.Response {
	t.Helper()

	var resp alice.Response
	require.NoError(t, json.Unmarshal(data, &resp))

	return resp
}

func newSessionRequest() alice.WebhookRequest {
	return alice.WebhookRequest{
		Meta: alice.Meta{
			Locale:   "ru-RU",
			Timezone: "Europe/Moscow",
			ClientID: "test-client",
		},
		Session: alice.Session{
			MessageID: 1,
			SessionID: "session-1",
			SkillID:   "skill-1",
			User:      &alice.User{UserID: "user-1"},
			Application: &alice.Application{
				ApplicationID: "app-1",
			},
			New: true,
		},
		Request: alice.Utterance{
			Command:           "",
			OriginalUtterance: "",
			Type:              "SimpleUtterance",
		},
		Version: alice.ProtocolVersion,
	}
}

func repeatRequest(command string) alice.WebhookRequest {
	req := newSessionRequest()
	req.Session.MessageID = 2
	req.Session.New = false
	req.Request.Command = command
	req.Request.OriginalUtterance = command

	return req
}

func TestWebhookRejectsMissingSecret(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, testSecret)
	status, body := doPost(t, app, "", marshalRequest(t, newSessionRequest()))

	assert.Equal(t, fiber.StatusUnauthorized, status)

	var errResp fiberfx.ErrorResponse
	require.NoError(t, json.Unmarshal(body, &errResp))
	assert.Equal(t, "unauthorized", errResp.Message)
	assert.Equal(t, fiber.StatusUnauthorized, errResp.Code)
}

func TestWebhookRejectsInvalidSecret(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, testSecret)
	status, _ := doPost(t, app, "wrong-secret", marshalRequest(t, newSessionRequest()))

	assert.Equal(t, fiber.StatusUnauthorized, status)
}

func TestWebhookRejectsWhenSecretIsNotConfigured(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, "")
	status, _ := doPost(t, app, "", marshalRequest(t, newSessionRequest()))

	assert.Equal(t, fiber.StatusUnauthorized, status)
}

func TestWebhookGreetsNewSession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, testSecret)
	status, body := doPost(t, app, testSecret, marshalRequest(t, newSessionRequest()))
	require.Equal(t, fiber.StatusOK, status)

	resp := decodeResponse(t, body)
	assert.Equal(t, alice.ProtocolVersion, resp.Version)
	assert.Equal(t, int64(1), resp.Session.MessageID)
	assert.Equal(t, "session-1", resp.Session.SessionID)
	assert.False(t, resp.Response.EndSession)
	assert.Contains(t, resp.Response.Text, "Hello")
	assert.NotEmpty(t, resp.Response.TTS)
	assert.NotContains(t, resp.Response.TTS, "<")
	assert.LessOrEqual(t, len([]rune(resp.Response.Text)), alice.MaxTextLength)
}

func TestWebhookEchoesRepeatUtterance(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, testSecret)
	status, body := doPost(t, app, testSecret, marshalRequest(t, repeatRequest("open the door")))
	require.Equal(t, fiber.StatusOK, status)

	resp := decodeResponse(t, body)
	assert.Equal(t, alice.ProtocolVersion, resp.Version)
	assert.False(t, resp.Response.EndSession)
	assert.Equal(t, "You said: open the door", resp.Response.Text)
	assert.Equal(t, "You said: open the door", resp.Response.TTS)
}

func TestWebhookAnswersAnonymousSession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, testSecret)

	body := `{
		"session": {"message_id": 7, "session_id": "anonymous-1", "new": false},
		"user_id": "top-level-must-be-ignored",
		"request": {"command": "hello there", "original_utterance": "hello there", "type": "SimpleUtterance"},
		"version": "1.0"
	}`
	status, data := doPost(t, app, testSecret, body)
	require.Equal(t, fiber.StatusOK, status)

	resp := decodeResponse(t, data)
	assert.Equal(t, "You said: hello there", resp.Response.Text)
}

func TestWebhookGreetsAnonymousNewSession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, testSecret)

	body := `{
		"session": {"message_id": 8, "session_id": "anonymous-2", "new": true},
		"request": {"command": "", "original_utterance": "", "type": "SimpleUtterance"},
		"version": "1.0"
	}`
	status, data := doPost(t, app, testSecret, body)
	require.Equal(t, fiber.StatusOK, status)

	resp := decodeResponse(t, data)
	assert.Contains(t, resp.Response.Text, "Hello")
}

func TestWebhookFallsBackToOriginalUtterance(t *testing.T) {
	t.Parallel()

	req := repeatRequest("")
	req.Request.OriginalUtterance = "spoken phrase"

	app := newTestApp(t, testSecret)
	status, body := doPost(t, app, testSecret, marshalRequest(t, req))
	require.Equal(t, fiber.StatusOK, status)

	resp := decodeResponse(t, body)
	assert.Equal(t, "You said: spoken phrase", resp.Response.Text)
}

func TestWebhookHandlesEmptyRepeatUtterance(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, testSecret)
	status, body := doPost(t, app, testSecret, marshalRequest(t, repeatRequest("")))
	require.Equal(t, fiber.StatusOK, status)

	resp := decodeResponse(t, body)
	assert.NotEmpty(t, resp.Response.Text)
	assert.LessOrEqual(t, len([]rune(resp.Response.Text)), alice.MaxTextLength)
}

func TestWebhookTruncatesLongTextAndIdentifiers(t *testing.T) {
	t.Parallel()

	req := repeatRequest(strings.Repeat("a", 2000))
	req.Session.SessionID = strings.Repeat("s", 100)
	req.Session.MessageID = 123456789012

	app := newTestApp(t, testSecret)
	status, body := doPost(t, app, testSecret, marshalRequest(t, req))
	require.Equal(t, fiber.StatusOK, status)

	resp := decodeResponse(t, body)
	assert.Len(t, []rune(resp.Response.Text), alice.MaxTextLength)
	assert.Len(t, []rune(resp.Response.TTS), alice.MaxTTSLength)
	assert.Len(t, []rune(resp.Session.SessionID), alice.MaxSessionIDLength)
	assert.Equal(t, int64(99_999_999), resp.Session.MessageID)
}

func TestWebhookRouteOnlyServesAlicePath(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, testSecret)

	unmountedStatus, _ := doPostPath(t, app, testSecret, "/api/v1/webhook", marshalRequest(t, newSessionRequest()))
	assert.Equal(t, fiber.StatusNotFound, unmountedStatus)

	aliceStatus, _ := doPostPath(t, app, testSecret, "/api/v1/alice/webhook", marshalRequest(t, newSessionRequest()))
	require.Equal(t, fiber.StatusOK, aliceStatus)
}

func TestWebhookNeverLogsSecret(t *testing.T) {
	t.Parallel()

	core, logs := observer.New(zapcore.WarnLevel)
	app := newTestAppWithLogger(t, testSecret, zap.New(core))

	const submittedSecret = "non-matching-secret"
	status, _ := doPost(t, app, submittedSecret, marshalRequest(t, newSessionRequest()))
	require.Equal(t, fiber.StatusUnauthorized, status)

	for _, entry := range logs.All() {
		assert.NotContains(t, entry.Message, testSecret)
		assert.NotContains(t, entry.Message, submittedSecret)
		for _, field := range entry.Context {
			assert.NotContains(t, field.String, testSecret)
			assert.NotContains(t, field.String, submittedSecret)
		}
	}
}

func TestWebhookRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, testSecret)
	status, body := doPost(t, app, testSecret, "{")

	assert.Equal(t, fiber.StatusBadRequest, status)

	var errResp fiberfx.ErrorResponse
	require.NoError(t, json.Unmarshal(body, &errResp))
	assert.Contains(t, errResp.Message, "failed to parse body")
	assert.Equal(t, fiber.StatusBadRequest, errResp.Code)
}
