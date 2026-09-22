package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type CreateAPIKeyInput struct {
	ProjectID uuid.UUID
	Name      string
}

type APIKeyService struct {
	repo    APIKeyRepository
	project ProjectVerifier
}

func NewAPIKeyService(repo APIKeyRepository, project ProjectVerifier) *APIKeyService {
	return &APIKeyService{repo: repo, project: project}
}

func (s *APIKeyService) CreateAPIKey(ctx context.Context, ownerID uuid.UUID, input CreateAPIKeyInput) (*APIKey, string, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, "", fmt.Errorf("%w: name is required", ErrValidation)
	}
	if len([]rune(input.Name)) > 100 {
		return nil, "", fmt.Errorf("%w: name must not exceed 100 characters", ErrValidation)
	}

	if err := s.project.VerifyProjectOwnership(ctx, input.ProjectID, ownerID); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return nil, "", ErrProjectNotFound
		}
		return nil, "", err
	}

	rawKey, keyPrefix, keyHash, err := generateAPIKey()
	if err != nil {
		return nil, "", fmt.Errorf("generate api key: %w", err)
	}

	key, err := s.repo.Create(ctx, CreateAPIKeyParams{
		ProjectID: input.ProjectID,
		OwnerID:   ownerID,
		Name:      input.Name,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
	})
	if err != nil {
		return nil, "", err
	}

	return key, rawKey, nil
}

func (s *APIKeyService) ListAPIKeys(ctx context.Context, ownerID uuid.UUID, projectID *uuid.UUID) ([]*APIKey, error) {
	if projectID != nil {
		if err := s.project.VerifyProjectOwnership(ctx, *projectID, ownerID); err != nil {
			if errors.Is(err, ErrProjectNotFound) {
				return nil, ErrProjectNotFound
			}
			return nil, err
		}
		return s.repo.ListByProject(ctx, *projectID, ownerID)
	}
	return s.repo.ListByOwner(ctx, ownerID)
}

func (s *APIKeyService) RevokeAPIKey(ctx context.Context, id, ownerID uuid.UUID) error {
	_, err := s.repo.Revoke(ctx, id, ownerID)
	return err
}

func (s *APIKeyService) ValidateKey(ctx context.Context, rawKey string) (*APIKey, error) {
	sum := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(sum[:])
	return s.repo.GetByHash(ctx, keyHash)
}

func generateAPIKey() (rawKey, keyPrefix, keyHash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("rand read: %w", err)
	}
	hexStr := hex.EncodeToString(b)
	rawKey = "lsk_" + hexStr
	keyPrefix = rawKey[:12]
	sum := sha256.Sum256([]byte(rawKey))
	keyHash = hex.EncodeToString(sum[:])
	return rawKey, keyPrefix, keyHash, nil
}
