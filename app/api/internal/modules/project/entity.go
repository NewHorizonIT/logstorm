package project

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	Slug        string
	Description string
	Environment string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
