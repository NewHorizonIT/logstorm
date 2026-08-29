package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/config"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewTokenService(cfg config.AuthConfig) *TokenService {
	return &TokenService{
		secretKey:       []byte(cfg.JWTSecret),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

func (s *TokenService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *TokenService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrTokenInvalid
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

func (s *TokenService) GenerateRefreshToken() (rawToken, hashedToken string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	rawToken = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(rawToken))
	hashedToken = hex.EncodeToString(sum[:])
	return rawToken, hashedToken, nil
}

func (s *TokenService) RefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}
