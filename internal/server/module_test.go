package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/capcom6/alice-companion/internal/server"
	"github.com/go-core-fx/fiberfx"
	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLogDoesNotContainQuerySecret(t *testing.T) {
	t.Parallel()

	const secret = "super-secret-key"

	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	app := fiberfx.New(
		fiberfx.Config{Address: "", ProxyHeader: "", Proxies: nil},
		server.NewFiberOptions(logger),
		prometheus.NewRegistry(),
		logger,
	)
	app.Get("/probe", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/probe?key="+secret, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	entries := logs.All()
	require.NotEmpty(t, entries, "request log must be emitted")

	var hasPath bool
	for _, entry := range entries {
		for _, field := range entry.Context {
			assert.NotContains(t, field.String, secret)
			if field.Key == "path" {
				hasPath = true
				assert.Equal(t, "/probe", field.String)
			}
		}
	}
	assert.True(t, hasPath, "request log must include the path field")
}
