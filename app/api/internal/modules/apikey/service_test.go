package apikey_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/apikey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fake repo ---

type fakeRepo struct {
	createFn        func(ctx context.Context, params apikey.CreateAPIKeyParams) (*apikey.APIKey, error)
	getByHashFn     func(ctx context.Context, keyHash string) (*apikey.APIKey, error)
	listByOwnerFn   func(ctx context.Context, ownerID uuid.UUID) ([]*apikey.APIKey, error)
	listByProjectFn func(ctx context.Context, projectID, ownerID uuid.UUID) ([]*apikey.APIKey, error)
	revokeFn        func(ctx context.Context, id, ownerID uuid.UUID) (*apikey.APIKey, error)
}

func (f *fakeRepo) Create(ctx context.Context, p apikey.CreateAPIKeyParams) (*apikey.APIKey, error) {
	return f.createFn(ctx, p)
}
func (f *fakeRepo) GetByHash(ctx context.Context, keyHash string) (*apikey.APIKey, error) {
	return f.getByHashFn(ctx, keyHash)
}
func (f *fakeRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*apikey.APIKey, error) {
	return f.listByOwnerFn(ctx, ownerID)
}
func (f *fakeRepo) ListByProject(ctx context.Context, projectID, ownerID uuid.UUID) ([]*apikey.APIKey, error) {
	return f.listByProjectFn(ctx, projectID, ownerID)
}
func (f *fakeRepo) Revoke(ctx context.Context, id, ownerID uuid.UUID) (*apikey.APIKey, error) {
	return f.revokeFn(ctx, id, ownerID)
}

// --- fake ProjectVerifier ---

type fakeProjectVerifier struct {
	verifyFn func(ctx context.Context, projectID, ownerID uuid.UUID) error
}

func (f *fakeProjectVerifier) VerifyProjectOwnership(ctx context.Context, projectID, ownerID uuid.UUID) error {
	return f.verifyFn(ctx, projectID, ownerID)
}

// --- helpers ---

func newSvc(repo apikey.APIKeyRepository, pv apikey.ProjectVerifier) *apikey.APIKeyService {
	return apikey.NewAPIKeyService(repo, pv)
}

func okVerifier() *fakeProjectVerifier {
	return &fakeProjectVerifier{
		verifyFn: func(_ context.Context, _, _ uuid.UUID) error { return nil },
	}
}

func notFoundVerifier() *fakeProjectVerifier {
	return &fakeProjectVerifier{
		verifyFn: func(_ context.Context, _, _ uuid.UUID) error { return apikey.ErrProjectNotFound },
	}
}

func fixedKey(ownerID, projectID uuid.UUID, name string) *apikey.APIKey {
	return &apikey.APIKey{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		ProjectID: projectID,
		Name:      name,
		KeyPrefix: "lsk_a1b2c3d4",
		KeyHash:   "fakehash",
	}
}

// --- CreateAPIKey ---

func TestCreateAPIKey_Success(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	var captured apikey.CreateAPIKeyParams

	repo := &fakeRepo{
		createFn: func(_ context.Context, p apikey.CreateAPIKeyParams) (*apikey.APIKey, error) {
			captured = p
			return fixedKey(ownerID, projectID, p.Name), nil
		},
	}

	key, rawKey, err := newSvc(repo, okVerifier()).CreateAPIKey(context.Background(), ownerID, apikey.CreateAPIKeyInput{
		ProjectID: projectID,
		Name:      "Production SDK",
	})

	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.NotEmpty(t, rawKey)
	assert.True(t, strings.HasPrefix(rawKey, "lsk_"), "raw key must start with lsk_")
	assert.Equal(t, 68, len(rawKey), "lsk_ (4) + hex(32 bytes) (64) = 68 chars")
	assert.Equal(t, rawKey[:12], captured.KeyPrefix, "key_prefix must be first 12 chars of raw key")
	assert.NotEqual(t, rawKey, captured.KeyHash, "key_hash must differ from raw key")
	assert.NotEmpty(t, captured.KeyHash)
	assert.Equal(t, "Production SDK", captured.Name)
	assert.Equal(t, ownerID, captured.OwnerID)
	assert.Equal(t, projectID, captured.ProjectID)
}

func TestCreateAPIKey_EmptyName(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	_, _, err := newSvc(repo, okVerifier()).CreateAPIKey(context.Background(), uuid.New(), apikey.CreateAPIKeyInput{
		ProjectID: uuid.New(),
		Name:      "   ",
	})

	assert.ErrorIs(t, err, apikey.ErrValidation)
}

func TestCreateAPIKey_NameTooLong(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	_, _, err := newSvc(repo, okVerifier()).CreateAPIKey(context.Background(), uuid.New(), apikey.CreateAPIKeyInput{
		ProjectID: uuid.New(),
		Name:      strings.Repeat("a", 101),
	})

	assert.ErrorIs(t, err, apikey.ErrValidation)
}

func TestCreateAPIKey_NameAtMaxLength(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{
		createFn: func(_ context.Context, p apikey.CreateAPIKeyParams) (*apikey.APIKey, error) {
			return fixedKey(uuid.New(), uuid.New(), p.Name), nil
		},
	}
	_, _, err := newSvc(repo, okVerifier()).CreateAPIKey(context.Background(), uuid.New(), apikey.CreateAPIKeyInput{
		ProjectID: uuid.New(),
		Name:      strings.Repeat("a", 100),
	})

	require.NoError(t, err)
}

func TestCreateAPIKey_ProjectNotFound(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	_, _, err := newSvc(repo, notFoundVerifier()).CreateAPIKey(context.Background(), uuid.New(), apikey.CreateAPIKeyInput{
		ProjectID: uuid.New(),
		Name:      "My Key",
	})

	assert.ErrorIs(t, err, apikey.ErrProjectNotFound)
}

func TestCreateAPIKey_KeysAreUnique(t *testing.T) {
	t.Parallel()

	seen := make(map[string]bool)
	repo := &fakeRepo{
		createFn: func(_ context.Context, p apikey.CreateAPIKeyParams) (*apikey.APIKey, error) {
			return fixedKey(uuid.New(), uuid.New(), p.Name), nil
		},
	}
	svc := newSvc(repo, okVerifier())

	for range 10 {
		_, rawKey, err := svc.CreateAPIKey(context.Background(), uuid.New(), apikey.CreateAPIKeyInput{
			ProjectID: uuid.New(),
			Name:      "key",
		})
		require.NoError(t, err)
		assert.False(t, seen[rawKey], "raw key must be unique")
		seen[rawKey] = true
	}
}

// --- ListAPIKeys ---

func TestListAPIKeys_AllProjects(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	want := []*apikey.APIKey{
		fixedKey(ownerID, uuid.New(), "A"),
		fixedKey(ownerID, uuid.New(), "B"),
	}
	repo := &fakeRepo{
		listByOwnerFn: func(_ context.Context, oID uuid.UUID) ([]*apikey.APIKey, error) {
			assert.Equal(t, ownerID, oID)
			return want, nil
		},
	}

	got, err := newSvc(repo, okVerifier()).ListAPIKeys(context.Background(), ownerID, nil)

	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestListAPIKeys_FilterByProject(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	want := []*apikey.APIKey{fixedKey(ownerID, projectID, "A")}
	repo := &fakeRepo{
		listByProjectFn: func(_ context.Context, pID, oID uuid.UUID) ([]*apikey.APIKey, error) {
			assert.Equal(t, projectID, pID)
			assert.Equal(t, ownerID, oID)
			return want, nil
		},
	}

	got, err := newSvc(repo, okVerifier()).ListAPIKeys(context.Background(), ownerID, &projectID)

	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestListAPIKeys_ProjectNotFound(t *testing.T) {
	t.Parallel()

	projectID := uuid.New()
	repo := &fakeRepo{}

	_, err := newSvc(repo, notFoundVerifier()).ListAPIKeys(context.Background(), uuid.New(), &projectID)

	assert.ErrorIs(t, err, apikey.ErrProjectNotFound)
}

// --- RevokeAPIKey ---

func TestRevokeAPIKey_Success(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	keyID := uuid.New()
	repo := &fakeRepo{
		revokeFn: func(_ context.Context, id, oID uuid.UUID) (*apikey.APIKey, error) {
			assert.Equal(t, keyID, id)
			assert.Equal(t, ownerID, oID)
			return fixedKey(ownerID, uuid.New(), "key"), nil
		},
	}

	err := newSvc(repo, okVerifier()).RevokeAPIKey(context.Background(), keyID, ownerID)

	require.NoError(t, err)
}

func TestRevokeAPIKey_NotFound(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{
		revokeFn: func(_ context.Context, _, _ uuid.UUID) (*apikey.APIKey, error) {
			return nil, apikey.ErrKeyNotFound
		},
	}

	err := newSvc(repo, okVerifier()).RevokeAPIKey(context.Background(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, apikey.ErrKeyNotFound)
}
