package config_test

import (
	"strings"
	"testing"

	"github.com/capcom6/alice-companion/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseURLDefaultsToSQLite(t *testing.T) {
	t.Parallel()

	assert.True(
		t,
		strings.HasPrefix(config.Default().Database.URL, "sqlite://"),
		config.Default().Database.URL,
	)
}

func TestDatabaseURLFromEnv(t *testing.T) {
	t.Setenv("DATABASE__URL", "sqlite:///var/lib/companion.db")

	cfg, err := config.New()
	require.NoError(t, err)

	assert.Equal(t, "sqlite:///var/lib/companion.db", cfg.Database.URL)
}
