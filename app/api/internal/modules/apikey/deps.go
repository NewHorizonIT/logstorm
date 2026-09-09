package apikey

import (
	"context"

	"github.com/google/uuid"
)

// ProjectVerifier là narrow interface apikey cần từ project domain.
// Satisfied by *project.ProjectService tại bootstrap.
type ProjectVerifier interface {
	VerifyProjectOwnership(ctx context.Context, projectID, ownerID uuid.UUID) error
}
