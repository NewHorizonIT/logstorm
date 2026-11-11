package search

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/project"
)

type SearchService struct {
	logRepo     LogRepository
	projectRepo project.ProjectRepository
}

func NewSearchService(logRepo LogRepository, projectRepo project.ProjectRepository) *SearchService {
	return &SearchService{logRepo: logRepo, projectRepo: projectRepo}
}

var ErrForbidden = errors.New("forbidden")

func (s *SearchService) SearchLogs(ctx context.Context, ownerID uuid.UUID, q *SearchQuery) (*SearchResult, error) {
	if _, err := s.projectRepo.GetByID(ctx, q.ProjectID, ownerID); err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	if err := q.Validate(); err != nil {
		return nil, err
	}

	return s.logRepo.Search(ctx, q)
}
