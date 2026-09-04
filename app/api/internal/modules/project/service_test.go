package project_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fake repo ---

type fakeProjectRepo struct {
	createFn            func(ctx context.Context, params project.CreateProjectParams) (*project.Project, error)
	getByIDFn           func(ctx context.Context, id, ownerID uuid.UUID) (*project.Project, error)
	getByOwnerAndSlugFn func(ctx context.Context, ownerID uuid.UUID, slug string) (*project.Project, error)
	listByOwnerFn       func(ctx context.Context, ownerID uuid.UUID) ([]*project.Project, error)
	updateFn            func(ctx context.Context, params project.UpdateProjectParams) (*project.Project, error)
}

func (f *fakeProjectRepo) Create(ctx context.Context, params project.CreateProjectParams) (*project.Project, error) {
	return f.createFn(ctx, params)
}
func (f *fakeProjectRepo) GetByID(ctx context.Context, id, ownerID uuid.UUID) (*project.Project, error) {
	return f.getByIDFn(ctx, id, ownerID)
}
func (f *fakeProjectRepo) GetByOwnerAndSlug(ctx context.Context, ownerID uuid.UUID, slug string) (*project.Project, error) {
	return f.getByOwnerAndSlugFn(ctx, ownerID, slug)
}
func (f *fakeProjectRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*project.Project, error) {
	return f.listByOwnerFn(ctx, ownerID)
}
func (f *fakeProjectRepo) Update(ctx context.Context, params project.UpdateProjectParams) (*project.Project, error) {
	return f.updateFn(ctx, params)
}

// --- helpers ---

func newSvc(repo project.ProjectRepository) *project.ProjectService {
	return project.NewProjectService(repo)
}

func fixedProject(ownerID uuid.UUID, name, slug string) *project.Project {
	return &project.Project{
		ID:          uuid.New(),
		OwnerID:     ownerID,
		Name:        name,
		Slug:        slug,
		Environment: "production",
	}
}

// --- CreateProject ---

func TestCreateProject_AutoGeneratesSlug(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	var captured project.CreateProjectParams
	repo := &fakeProjectRepo{
		createFn: func(_ context.Context, p project.CreateProjectParams) (*project.Project, error) {
			captured = p
			return fixedProject(ownerID, p.Name, p.Slug), nil
		},
	}

	_, err := newSvc(repo).CreateProject(context.Background(), ownerID, project.CreateProjectInput{
		Name: "My Awesome App",
	})

	require.NoError(t, err)
	assert.Equal(t, "my-awesome-app", captured.Slug)
}

func TestCreateProject_ExplicitSlug(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	var captured project.CreateProjectParams
	repo := &fakeProjectRepo{
		createFn: func(_ context.Context, p project.CreateProjectParams) (*project.Project, error) {
			captured = p
			return fixedProject(ownerID, p.Name, p.Slug), nil
		},
	}

	_, err := newSvc(repo).CreateProject(context.Background(), ownerID, project.CreateProjectInput{
		Name: "My App",
		Slug: "custom-slug",
	})

	require.NoError(t, err)
	assert.Equal(t, "custom-slug", captured.Slug)
}

func TestCreateProject_DefaultsEnvironmentToProduction(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	var captured project.CreateProjectParams
	repo := &fakeProjectRepo{
		createFn: func(_ context.Context, p project.CreateProjectParams) (*project.Project, error) {
			captured = p
			return fixedProject(ownerID, p.Name, p.Slug), nil
		},
	}

	_, err := newSvc(repo).CreateProject(context.Background(), ownerID, project.CreateProjectInput{
		Name: "My App",
	})

	require.NoError(t, err)
	assert.Equal(t, "production", captured.Environment)
}

func TestCreateProject_EmptyName(t *testing.T) {
	t.Parallel()

	repo := &fakeProjectRepo{}
	_, err := newSvc(repo).CreateProject(context.Background(), uuid.New(), project.CreateProjectInput{
		Name: "   ",
	})
	assert.ErrorIs(t, err, project.ErrValidation)
}

func TestCreateProject_NameTooLong(t *testing.T) {
	t.Parallel()

	repo := &fakeProjectRepo{}
	_, err := newSvc(repo).CreateProject(context.Background(), uuid.New(), project.CreateProjectInput{
		Name: strings.Repeat("a", 101),
	})
	assert.ErrorIs(t, err, project.ErrValidation)
}

func TestCreateProject_InvalidSlug(t *testing.T) {
	t.Parallel()

	cases := []string{
		"My App",
		"-leading",
		"trailing-",
		"double--hyp",
		"UPPER",
		"has space",
	}

	for _, slug := range cases {
		slug := slug
		t.Run(slug, func(t *testing.T) {
			t.Parallel()
			repo := &fakeProjectRepo{}
			_, err := newSvc(repo).CreateProject(context.Background(), uuid.New(), project.CreateProjectInput{
				Name: "Valid Name",
				Slug: slug,
			})
			assert.ErrorIs(t, err, project.ErrValidation, "slug %q should fail validation", slug)
		})
	}
}

func TestCreateProject_InvalidEnvironment(t *testing.T) {
	t.Parallel()

	repo := &fakeProjectRepo{}
	_, err := newSvc(repo).CreateProject(context.Background(), uuid.New(), project.CreateProjectInput{
		Name:        "My App",
		Environment: "local",
	})
	assert.ErrorIs(t, err, project.ErrValidation)
}

func TestCreateProject_SlugAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := &fakeProjectRepo{
		createFn: func(_ context.Context, _ project.CreateProjectParams) (*project.Project, error) {
			return nil, project.ErrSlugAlreadyExists
		},
	}
	_, err := newSvc(repo).CreateProject(context.Background(), uuid.New(), project.CreateProjectInput{
		Name: "My App",
	})
	assert.ErrorIs(t, err, project.ErrSlugAlreadyExists)
}

// --- GetProject ---

func TestGetProject_Success(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	want := fixedProject(ownerID, "App", "app")
	repo := &fakeProjectRepo{
		getByIDFn: func(_ context.Context, id, oID uuid.UUID) (*project.Project, error) {
			assert.Equal(t, want.ID, id)
			assert.Equal(t, ownerID, oID)
			return want, nil
		},
	}

	got, err := newSvc(repo).GetProject(context.Background(), want.ID, ownerID)

	require.NoError(t, err)
	assert.Equal(t, want.ID, got.ID)
}

func TestGetProject_NotFound(t *testing.T) {
	t.Parallel()

	repo := &fakeProjectRepo{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*project.Project, error) {
			return nil, project.ErrProjectNotFound
		},
	}
	_, err := newSvc(repo).GetProject(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, project.ErrProjectNotFound)
}

// --- ListProjects ---

func TestListProjects_ReturnsList(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	want := []*project.Project{
		fixedProject(ownerID, "A", "a"),
		fixedProject(ownerID, "B", "b"),
	}
	repo := &fakeProjectRepo{
		listByOwnerFn: func(_ context.Context, oID uuid.UUID) ([]*project.Project, error) {
			assert.Equal(t, ownerID, oID)
			return want, nil
		},
	}

	got, err := newSvc(repo).ListProjects(context.Background(), ownerID)

	require.NoError(t, err)
	assert.Len(t, got, 2)
}

// --- UpdateProject ---

func TestUpdateProject_Success(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	var captured project.UpdateProjectParams
	repo := &fakeProjectRepo{
		updateFn: func(_ context.Context, p project.UpdateProjectParams) (*project.Project, error) {
			captured = p
			return fixedProject(ownerID, p.Name, "existing-slug"), nil
		},
	}

	_, err := newSvc(repo).UpdateProject(context.Background(), ownerID, project.UpdateProjectInput{
		ID:   projectID,
		Name: "New Name",
	})

	require.NoError(t, err)
	assert.Equal(t, projectID, captured.ID)
	assert.Equal(t, ownerID, captured.OwnerID)
	assert.Equal(t, "New Name", captured.Name)
}

func TestUpdateProject_EmptyName(t *testing.T) {
	t.Parallel()

	repo := &fakeProjectRepo{}
	_, err := newSvc(repo).UpdateProject(context.Background(), uuid.New(), project.UpdateProjectInput{
		ID:   uuid.New(),
		Name: "",
	})
	assert.ErrorIs(t, err, project.ErrValidation)
}

func TestUpdateProject_NotFound(t *testing.T) {
	t.Parallel()

	repo := &fakeProjectRepo{
		updateFn: func(_ context.Context, _ project.UpdateProjectParams) (*project.Project, error) {
			return nil, project.ErrProjectNotFound
		},
	}
	_, err := newSvc(repo).UpdateProject(context.Background(), uuid.New(), project.UpdateProjectInput{
		ID:   uuid.New(),
		Name: "Valid Name",
	})
	assert.ErrorIs(t, err, project.ErrProjectNotFound)
}

 func TestCreateProject_NameAtMaxLength(t *testing.T) {
      repo := &fakeProjectRepo{createFn: func(_ context.Context, p project.CreateProjectParams) (*project.Project, error) {
          return fixedProject(uuid.New(), p.Name, p.Slug), nil
      }}
      _, err := newSvc(repo).CreateProject(context.Background(), uuid.New(), project.CreateProjectInput{
          Name: strings.Repeat("a", 100),
      })
      require.NoError(t, err)
  }
