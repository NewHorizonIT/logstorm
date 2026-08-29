package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTokenService() *auth.TokenService {
	return auth.NewTokenService(config.AuthConfig{
		JWTSecret:       "test-secret-key-minimum-32-chars!!",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	})
}

func TestTokenService_GenerateAndValidateAccessToken(t *testing.T) {
	t.Parallel()

	svc := testTokenService()
	userID := uuid.New()

	token, err := svc.GenerateAccessToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestTokenService_ValidateAccessToken_Expired(t *testing.T) {
	t.Parallel()

	svc := auth.NewTokenService(config.AuthConfig{
		JWTSecret:       "test-secret-key-minimum-32-chars!!",
		AccessTokenTTL:  -time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	})
	userID := uuid.New()

	token, err := svc.GenerateAccessToken(userID)
	require.NoError(t, err)

	_, err = svc.ValidateAccessToken(token)
	assert.ErrorIs(t, err, auth.ErrTokenInvalid)
}

func TestTokenService_ValidateAccessToken_InvalidSignature(t *testing.T) {
	t.Parallel()

	svc := testTokenService()

	_, err := svc.ValidateAccessToken("invalid.jwt.token")
	assert.ErrorIs(t, err, auth.ErrTokenInvalid)
}

func TestTokenService_GenerateRefreshToken_Unique(t *testing.T) {
	t.Parallel()

	svc := testTokenService()

	raw1, hash1, err := svc.GenerateRefreshToken()
	require.NoError(t, err)
	assert.NotEmpty(t, raw1)
	assert.NotEmpty(t, hash1)
	assert.NotEqual(t, raw1, hash1)

	raw2, hash2, err := svc.GenerateRefreshToken()
	require.NoError(t, err)
	assert.NotEqual(t, raw1, raw2)
	assert.NotEqual(t, hash1, hash2)
}

func TestTokenService_RefreshTokenTTL(t *testing.T) {
	t.Parallel()

	svc := testTokenService()
	assert.Equal(t, 7*24*time.Hour, svc.RefreshTokenTTL())
}
