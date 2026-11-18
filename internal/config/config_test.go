package config_test

import (
	"testing"

	"github.com/capcom6/alice-companion/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAliceSharedSecretFromEnv(t *testing.T) {
	t.Setenv("ALICE__SHARED_SECRET", "env-secret")

	cfg, err := config.New()
	require.NoError(t, err)

	assert.Equal(t, "env-secret", cfg.Alice.SharedSecret)
}

func TestAliceSharedSecretDefaultsToEmpty(t *testing.T) {
	t.Setenv("ALICE__SHARED_SECRET", "")

	cfg, err := config.New()
	require.NoError(t, err)

	assert.Empty(t, cfg.Alice.SharedSecret)
}
