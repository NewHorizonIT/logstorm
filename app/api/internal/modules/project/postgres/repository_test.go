package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logstorm/api/internal/modules/project"
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
	os.Exit(run(m))
}

func run(m *testing.M) int {
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

	return m.Run()
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

func newRepo() *PostgresProjectRepository {
	return NewPostgresProjectRepository(testPool)
}

func createTestUser(t *testing.T) *user.User {
	t.Helper()
	repo := userpostgres.NewPostgresUserRepository(testPool)
	u, err := repo.Create(context.Background(), user.CreateUserParams{
		Email:        fmt.Sprintf("proj-test-%s@example.com", uuid.New()),
		PasswordHash: "hashed",
		FullName:     "Project Test User",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		testPool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", u.ID)
	})
	return u
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	proj, err := repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u.ID,
		Name:        "My Project",
		Slug:        "my-project",
		Description: "A test project",
		Environment: "production",
	})

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, proj.ID)
	assert.Equal(t, u.ID, proj.OwnerID)
	assert.Equal(t, "My Project", proj.Name)
	assert.Equal(t, "my-project", proj.Slug)
	assert.Equal(t, "A test project", proj.Description)
	assert.Equal(t, "production", proj.Environment)
	assert.False(t, proj.CreatedAt.IsZero())
}

func TestCreate_NoDescription(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	proj, err := repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u.ID,
		Name:        "No Desc",
		Slug:        fmt.Sprintf("no-desc-%s", uuid.New()),
		Environment: "staging",
	})

	require.NoError(t, err)
	assert.Equal(t, "", proj.Description)
}

func TestCreate_DuplicateSlug(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	slug := fmt.Sprintf("dup-slug-%s", uuid.New())
	_, err := repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u.ID,
		Name:        "First",
		Slug:        slug,
		Environment: "production",
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u.ID,
		Name:        "Second",
		Slug:        slug,
		Environment: "production",
	})

	assert.ErrorIs(t, err, project.ErrSlugAlreadyExists)
}

func TestCreate_SameSlugDifferentOwner(t *testing.T) {
	t.Parallel()

	u1 := createTestUser(t)
	u2 := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	slug := fmt.Sprintf("shared-%s", uuid.New())

	_, err := repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u1.ID,
		Name:        "App",
		Slug:        slug,
		Environment: "production",
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u2.ID,
		Name:        "App",
		Slug:        slug,
		Environment: "production",
	})

	require.NoError(t, err, "slug unique per owner — different owner must be allowed")
}

func TestGetByID_Found(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	created, err := repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u.ID,
		Name:        "Get Test",
		Slug:        fmt.Sprintf("get-test-%s", uuid.New()),
		Environment: "staging",
	})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, created.ID, u.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "staging", got.Environment)
}

func TestGetByID_NotFound(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	_, err := repo.GetByID(context.Background(), uuid.New(), uuid.New())

	assert.ErrorIs(t, err, project.ErrProjectNotFound)
}

func TestGetByID_WrongOwner(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	created, err := repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u.ID,
		Name:        "Owner Test",
		Slug:        fmt.Sprintf("owner-test-%s", uuid.New()),
		Environment: "production",
	})
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, created.ID, uuid.New())

	assert.ErrorIs(t, err, project.ErrProjectNotFound)
}

func TestGetByOwnerAndSlug_Found(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	slug := fmt.Sprintf("slug-lookup-%s", uuid.New())
	created, err := repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u.ID,
		Name:        "Slug Lookup",
		Slug:        slug,
		Environment: "production",
	})
	require.NoError(t, err)

	got, err := repo.GetByOwnerAndSlug(ctx, u.ID, slug)

	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

func TestGetByOwnerAndSlug_NotFound(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	_, err := repo.GetByOwnerAndSlug(context.Background(), uuid.New(), "nonexistent-slug")

	assert.ErrorIs(t, err, project.ErrProjectNotFound)
}

func TestListByOwner(t *testing.T) {
	t.Parallel()

	u1 := createTestUser(t)
	u2 := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	for i := range 3 {
		_, err := repo.Create(ctx, project.CreateProjectParams{
			OwnerID:     u1.ID,
			Name:        fmt.Sprintf("Project %d", i),
			Slug:        fmt.Sprintf("project-%d-%s", i, uuid.New()),
			Environment: "production",
		})
		require.NoError(t, err)
	}

	_, err := repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u2.ID,
		Name:        "Other User Project",
		Slug:        fmt.Sprintf("other-%s", uuid.New()),
		Environment: "production",
	})
	require.NoError(t, err)

	projects, err := repo.ListByOwner(ctx, u1.ID)

	require.NoError(t, err)
	assert.Len(t, projects, 3)
	for _, p := range projects {
		assert.Equal(t, u1.ID, p.OwnerID)
	}
}

func TestUpdate_Success(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	created, err := repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u.ID,
		Name:        "Old Name",
		Slug:        fmt.Sprintf("update-test-%s", uuid.New()),
		Environment: "production",
	})
	require.NoError(t, err)

	updated, err := repo.Update(ctx, project.UpdateProjectParams{
		ID:          created.ID,
		OwnerID:     u.ID,
		Name:        "New Name",
		Description: "Now has a description",
	})

	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, "Now has a description", updated.Description)
	assert.Equal(t, created.Slug, updated.Slug, "slug does not change on update")
}

func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	_, err := repo.Update(context.Background(), project.UpdateProjectParams{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Name:    "Ghost",
	})

	assert.ErrorIs(t, err, project.ErrProjectNotFound)
}

func TestUpdate_WrongOwner(t *testing.T) {
	t.Parallel()

	u := createTestUser(t)
	repo := newRepo()
	ctx := context.Background()

	created, err := repo.Create(ctx, project.CreateProjectParams{
		OwnerID:     u.ID,
		Name:        "Protected",
		Slug:        fmt.Sprintf("protected-%s", uuid.New()),
		Environment: "production",
	})
	require.NoError(t, err)

	_, err = repo.Update(ctx, project.UpdateProjectParams{
		ID:      created.ID,
		OwnerID: uuid.New(),
		Name:    "Hacked",
	})

	assert.ErrorIs(t, err, project.ErrProjectNotFound)
}
