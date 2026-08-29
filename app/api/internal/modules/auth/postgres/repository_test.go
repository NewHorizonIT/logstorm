package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/modules/user"
	userpostgres "github.com/logstorm/api/internal/modules/user/postgres"
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
	migrationsDir := filepath.Clean(
		filepath.Join(filepath.Dir(file), "../../../../migrations"),
	)

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

func tokenHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}

func createTestUser(t *testing.T) *user.User {
	t.Helper()
	repo := userpostgres.NewPostgresUserRepository(testPool)
	u, err := repo.Create(context.Background(), user.CreateUserParams{
		Email:        fmt.Sprintf("auth-test-%s@example.com", uuid.New()),
		PasswordHash: "hashed",
		FullName:     "Auth Test User",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		testPool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", u.ID)
	})
	return u
}

func newRepo() *PostgresAuthRepository {
	return NewPostgresAuthRepository(testPool)
}

func TestCreateRefreshToken(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	hash := tokenHash("rawtoken123")
	token, err := repo.CreateRefreshToken(ctx, auth.CreateRefreshTokenParams{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	require.NoError(t, err)
	assert.Equal(t, u.ID, token.UserID)
	assert.Equal(t, hash, token.TokenHash)
	assert.Nil(t, token.RevokedAt)
	assert.False(t, token.CreatedAt.IsZero())
}

func TestGetRefreshTokenByHash_Found(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	hash := tokenHash("findme")
	_, err := repo.CreateRefreshToken(ctx, auth.CreateRefreshTokenParams{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	require.NoError(t, err)

	found, err := repo.GetRefreshTokenByHash(ctx, hash)

	require.NoError(t, err)
	assert.Equal(t, hash, found.TokenHash)
}

func TestGetRefreshTokenByHash_NotFound(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	_, err := repo.GetRefreshTokenByHash(context.Background(), "nonexistent-hash")

	assert.ErrorIs(t, err, auth.ErrRefreshTokenNotFound)
}

func TestRevokeRefreshToken(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	hash := tokenHash("revoke-me")
	_, err := repo.CreateRefreshToken(ctx, auth.CreateRefreshTokenParams{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	require.NoError(t, err)

	err = repo.RevokeRefreshToken(ctx, hash)
	require.NoError(t, err)

	_, err = repo.GetRefreshTokenByHash(ctx, hash)
	assert.ErrorIs(t, err, auth.ErrRefreshTokenNotFound)
}

func TestDeleteExpiredTokens(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	hash := tokenHash("already-expired")
	_, err := repo.CreateRefreshToken(ctx, auth.CreateRefreshTokenParams{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	})
	require.NoError(t, err)

	err = repo.DeleteExpiredTokens(ctx)
	require.NoError(t, err)

	_, err = repo.GetRefreshTokenByHash(ctx, hash)
	assert.ErrorIs(t, err, auth.ErrRefreshTokenNotFound)
}
