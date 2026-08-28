package postgres

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logstorm/api/internal/modules/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainerspg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := testcontainerspg.Run(ctx,
		"postgres:16",
		testcontainerspg.WithDatabase("logstorm_test"),
		testcontainerspg.WithUsername("test"),
		testcontainerspg.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			panic("failed to terminate postgres container: " + err.Error())
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic("failed to create pool: " + err.Error())
	}
	defer testPool.Close()

	if err := applySchema(ctx, testPool); err != nil {
		panic("failed to apply schema: " + err.Error())
	}

	m.Run()
}

func applySchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, file, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../migrations"))

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		sql, err := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return err
		}
	}
	return nil
}

func newRepo() *PostgresUserRepository {
	return NewPostgresUserRepository(testPool)
}

func cleanupUser(t *testing.T, id uuid.UUID) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := testPool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", id); err != nil {
			t.Logf("cleanup: failed to delete user %s: %v", id, err)
		}
	})
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()

	repo := newRepo()
	ctx := context.Background()

	got, err := repo.Create(ctx, user.CreateUserParams{
		Email:        "create_success@example.com",
		PasswordHash: "hashed",
		FullName:     "Test User",
	})

	require.NoError(t, err)
	cleanupUser(t, got.ID)

	assert.NotEqual(t, uuid.Nil, got.ID)
	assert.Equal(t, "create_success@example.com", got.Email)
	assert.Equal(t, "Test User", got.FullName)
	assert.Equal(t, "active", got.Status)
	assert.False(t, got.IsVerified)
	assert.Nil(t, got.LastLoginAt)
	assert.False(t, got.CreatedAt.IsZero())
}

func TestCreate_DuplicateEmail(t *testing.T) {
	t.Parallel()

	repo := newRepo()
	ctx := context.Background()

	first, err := repo.Create(ctx, user.CreateUserParams{
		Email:        "duplicate@example.com",
		PasswordHash: "hashed",
		FullName:     "First User",
	})
	require.NoError(t, err)
	cleanupUser(t, first.ID)

	_, err = repo.Create(ctx, user.CreateUserParams{
		Email:        "duplicate@example.com",
		PasswordHash: "hashed",
		FullName:     "Second User",
	})

	assert.ErrorIs(t, err, user.ErrUserEmailAlreadyExists)
}

func TestCreate_EmailCaseInsensitiveUnique(t *testing.T) {
	t.Parallel()

	repo := newRepo()
	ctx := context.Background()

	first, err := repo.Create(ctx, user.CreateUserParams{
		Email:        "citext@example.com",
		PasswordHash: "hashed",
		FullName:     "First User",
	})
	require.NoError(t, err)
	cleanupUser(t, first.ID)

	_, err = repo.Create(ctx, user.CreateUserParams{
		Email:        "CITEXT@EXAMPLE.COM",
		PasswordHash: "hashed",
		FullName:     "Second User",
	})

	assert.ErrorIs(t, err, user.ErrUserEmailAlreadyExists)
}

func TestGetByID_Found(t *testing.T) {
	t.Parallel()

	repo := newRepo()
	ctx := context.Background()

	created, err := repo.Create(ctx, user.CreateUserParams{
		Email:        "getbyid@example.com",
		PasswordHash: "hashed",
		FullName:     "Get By ID",
	})
	require.NoError(t, err)
	cleanupUser(t, created.ID)

	got, err := repo.GetByID(ctx, created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "getbyid@example.com", got.Email)
}

func TestGetByID_NotFound(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	_, err := repo.GetByID(context.Background(), uuid.New())

	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestGetByEmail_Found(t *testing.T) {
	t.Parallel()

	repo := newRepo()
	ctx := context.Background()

	created, err := repo.Create(ctx, user.CreateUserParams{
		Email:        "getbyemail@example.com",
		PasswordHash: "hashed",
		FullName:     "Get By Email",
	})
	require.NoError(t, err)
	cleanupUser(t, created.ID)

	got, err := repo.GetByEmail(ctx, "GETBYEMAIL@EXAMPLE.COM")

	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

func TestGetByEmail_NotFound(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	_, err := repo.GetByEmail(context.Background(), "nobody@example.com")

	assert.ErrorIs(t, err, user.ErrUserNotFound)
}
