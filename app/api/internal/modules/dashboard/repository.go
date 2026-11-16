package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DashboardRepository interface {
	GetStats(ctx context.Context, projectID uuid.UUID, from, to time.Time) (*LogStats, error)
}
