package dashboard

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/project"
)

var (
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidTimeRange  = errors.New("invalid time range")
	ErrTimeRangeTooLarge = errors.New("time range too large")
)

type DashboardService struct {
	repo        DashboardRepository
	projectRepo project.ProjectRepository
}

func NewDashboardService(repo DashboardRepository, projectRepo project.ProjectRepository) *DashboardService {
	return &DashboardService{repo: repo, projectRepo: projectRepo}
}

func (s *DashboardService) GetStats(ctx context.Context, ownerID, projectID uuid.UUID, from, to time.Time) (*LogStats, error) {
	if _, err := s.projectRepo.GetByID(ctx, projectID, ownerID); err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	if from.IsZero() {
		from = time.Now().UTC().Add(-24 * time.Hour)
	}
	if to.IsZero() {
		to = time.Now().UTC()
	}

	if from.After(to) {
		return nil, ErrInvalidTimeRange
	}

	if to.Sub(from) > 90*24*time.Hour {
		return nil, ErrTimeRangeTooLarge
	}

	return s.repo.GetStats(ctx, projectID, from, to)
}
