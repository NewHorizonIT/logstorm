package project

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	multiHyphen = regexp.MustCompile(`-{2,}`)
)

var validEnvironments = map[string]bool{
	"production":  true,
	"staging":     true,
	"development": true,
}

type CreateProjectInput struct {
	Name        string
	Slug        string
	Description string
	Environment string
}

type UpdateProjectInput struct {
	ID          uuid.UUID
	Name        string
	Description string
}

type ProjectService struct {
	repo ProjectRepository
}

func NewProjectService(repo ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) CreateProject(ctx context.Context, ownerID uuid.UUID, input CreateProjectInput) (*Project, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrValidation)
	}
	if len([]rune(input.Name)) > 100 {
		return nil, fmt.Errorf("%w: name must not exceed 100 characters", ErrValidation)
	}

	if input.Slug == "" {
		input.Slug = generateSlug(input.Name)
	}
	if !slugPattern.MatchString(input.Slug) {
		return nil, fmt.Errorf("%w: slug must contain only lowercase letters, digits, and hyphens, and must not start or end with a hyphen", ErrValidation)
	}

	if len(input.Slug) > 100 {
		return nil, fmt.Errorf("%w: slug must not exceed 100 characters", ErrValidation)
	}


	if input.Environment == "" {
		input.Environment = "production"
	}
	if !validEnvironments[input.Environment] {
		return nil, fmt.Errorf("%w: environment must be one of: production, staging, development", ErrValidation)
	}

	return s.repo.Create(ctx, CreateProjectParams{
		OwnerID:     ownerID,
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		Environment: input.Environment,
	})
}

func (s *ProjectService) GetProject(ctx context.Context, id, ownerID uuid.UUID) (*Project, error) {
	return s.repo.GetByID(ctx, id, ownerID)
}

func (s *ProjectService) ListProjects(ctx context.Context, ownerID uuid.UUID) ([]*Project, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

func (s *ProjectService) UpdateProject(ctx context.Context, ownerID uuid.UUID, input UpdateProjectInput) (*Project, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrValidation)
	}
	if len([]rune(input.Name)) > 100 {
		return nil, fmt.Errorf("%w: name must not exceed 100 characters", ErrValidation)
	}

	return s.repo.Update(ctx, UpdateProjectParams{
		ID:          input.ID,
		OwnerID:     ownerID,
		Name:        input.Name,
		Description: input.Description,
	})
}

func generateSlug(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	slug := strings.Trim(multiHyphen.ReplaceAllString(b.String(), "-"), "-")
	if slug == "" {
		slug = "project"
	}
	if len(slug) > 63 {
		slug = strings.TrimRight(slug[:63], "-")
	}
	return slug
}
