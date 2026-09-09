package apikey

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	KeyHash   string
	KeyPrefix string
	CreatedAt time.Time
	RevokedAt *time.Time
}
