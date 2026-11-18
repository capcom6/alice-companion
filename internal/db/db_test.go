package db_test

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/capcom6/alice-companion/internal/db"
	"github.com/capcom6/alice-companion/internal/db/migrations"
	"github.com/go-core-fx/goosefx"
	"github.com/go-core-fx/sqlfx"
	"github.com/pressly/goose/v3/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/schema"
	"go.uber.org/fx"
)

// memoryDSN is the in-memory SQLite DSN used by the repository tests.
const memoryDSN = "sqlite://:memory:"

// testItem is a throwaway model used to exercise the bun + SQLite driver path.
type testItem struct {
	bun.BaseModel `bun:"table:test_items,alias:ti"`

	ID        int64     `bun:",pk,autoincrement"`
	Value     string    `bun:"value,notnull"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull"`
}

// openSQL opens a database through the same sqlfx path production uses.
func openSQL(t *testing.T, dsn string) *sql.DB {
	t.Helper()

	sqlDB, err := sqlfx.New(sqlfx.Config{
		URL:             dsn,
		ConnMaxIdleTime: 0,
		ConnMaxLifetime: 0,
		MaxOpenConns:    1,
		MaxIdleConns:    1,
	})
	require.NoError(t, err)

	t.Cleanup(func() { _ = sqlDB.Close() })

	return sqlDB
}

// applyMigrations runs the embedded goose migrations and returns how many ran.
func applyMigrations(t *testing.T, sqlDB *sql.DB) int {
	t.Helper()

	provider, err := goosefx.NewProvider(sqlDB, goosefx.Storage(migrations.FS), database.DialectSQLite3)
	require.NoError(t, err)

	res, err := provider.Up(context.Background())
	require.NoError(t, err)

	return len(res)
}

func TestSQLiteDriverPathAcceptsDSNVariants(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	cwd, err := os.Getwd()
	require.NoError(t, err)

	relPath, err := filepath.Rel(cwd, filepath.Join(dir, "rel.db"))
	require.NoError(t, err)

	tests := []struct {
		name string
		dsn  string
	}{
		{name: "memory", dsn: memoryDSN},
		{name: "relative file", dsn: "sqlite://" + relPath},
		{name: "absolute file", dsn: "sqlite://" + filepath.Join(dir, "abs.db")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sqlDB := openSQL(t, tt.dsn)
			require.NoError(t, sqlDB.PingContext(context.Background()))
		})
	}
}

func TestSQLiteDriverPathRejectsUnsupportedScheme(t *testing.T) {
	t.Parallel()

	_, err := sqlfx.New(sqlfx.Config{URL: "oracle://user:pass@localhost:1521/db"})
	require.Error(t, err)
}

func TestMigrationsApplyOnFreshAndExistingDB(t *testing.T) {
	t.Parallel()

	dsn := "sqlite://" + filepath.Join(t.TempDir(), "existing.db")

	fresh := openSQL(t, dsn)
	assert.Positive(t, applyMigrations(t, fresh), "fresh DB must apply the embedded migrations")
	require.NoError(t, fresh.Close())

	existing := openSQL(t, dsn)
	assert.Zero(t, applyMigrations(t, existing), "existing DB must be a no-op")
	require.NoError(t, existing.Close())
}

func TestBunSQLiteRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	sqlDB := openSQL(t, memoryDSN)
	bunDB := bun.NewDB(sqlDB, sqlitedialect.New())
	t.Cleanup(func() { _ = bunDB.Close() })

	_, err := bunDB.NewCreateTable().Model((*testItem)(nil)).IfNotExists().Exec(ctx)
	require.NoError(t, err)

	now := time.Now().UTC()
	items := []*testItem{
		{Value: "alpha", CreatedAt: now},
		{Value: "beta", CreatedAt: now},
	}
	_, err = bunDB.NewInsert().Model(&items).Exec(ctx)
	require.NoError(t, err)

	var loaded []testItem
	err = bunDB.NewSelect().
		Model(&loaded).
		Where("value IN (?)", bun.List([]string{"alpha", "beta"})).
		Order("id ASC").
		Scan(ctx)
	require.NoError(t, err)

	require.Len(t, loaded, 2)
	assert.Equal(t, "alpha", loaded[0].Value)
	assert.Equal(t, "beta", loaded[1].Value)
	assert.WithinDuration(t, now, loaded[0].CreatedAt, time.Second)
}

func TestModuleProvidesSQLiteDialects(t *testing.T) {
	t.Parallel()

	var (
		schemaDialect schema.Dialect
		gooseDialect  database.Dialect
		storage       goosefx.Storage
	)

	app := fx.New(
		db.Module(),
		fx.NopLogger,
		fx.Populate(&schemaDialect, &gooseDialect, &storage),
	)
	require.NoError(t, app.Err())

	assert.Equal(t, dialect.SQLite, schemaDialect.Name())
	assert.Equal(t, database.DialectSQLite3, gooseDialect)
	_, err := fs.Stat(storage, "20260204013005_initial.sql")
	assert.NoError(t, err)
}
