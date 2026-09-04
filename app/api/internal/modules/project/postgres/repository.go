package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logstorm/api/internal/db"
	"github.com/logstorm/api/internal/modules/project"
)

const uniqueViolationCode = "23505"

type PostgresProjectRepository struct {
	q *db.Queries
}

func NewPostgresProjectRepository(pool *pgxpool.Pool) *PostgresProjectRepository {
	return &PostgresProjectRepository{q: db.New(pool)}
}

func (r *PostgresProjectRepository) Create(ctx context.Context, params project.CreateProjectParams) (*project.Project, error) {
	row, err := r.q.CreateProject(ctx, db.CreateProjectParams{
		OwnerID:     pgtype.UUID{Bytes: params.OwnerID, Valid: true},
		Name:        params.Name,
		Slug:        params.Slug,
		Description: pgtype.Text{String: params.Description, Valid: params.Description != ""},
		Environment: params.Environment,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return nil, project.ErrSlugAlreadyExists
		}
		return nil, err
	}
	return toProject(row), nil
}

func (r *PostgresProjectRepository) GetByID(ctx context.Context, id, ownerID uuid.UUID) (*project.Project, error) {
	row, err := r.q.GetProjectByID(ctx, db.GetProjectByIDParams{
		ID:      pgtype.UUID{Bytes: id, Valid: true},
		OwnerID: pgtype.UUID{Bytes: ownerID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, project.ErrProjectNotFound
		}
		return nil, err
	}
	return toProject(row), nil
}

func (r *PostgresProjectRepository) GetByOwnerAndSlug(ctx context.Context, ownerID uuid.UUID, slug string) (*project.Project, error) {
	row, err := r.q.GetProjectByOwnerAndSlug(ctx, db.GetProjectByOwnerAndSlugParams{
		OwnerID: pgtype.UUID{Bytes: ownerID, Valid: true},
		Slug:    slug,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, project.ErrProjectNotFound
		}
		return nil, err
	}
	return toProject(row), nil
}

func (r *PostgresProjectRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*project.Project, error) {
	rows, err := r.q.ListProjectsByOwner(ctx, pgtype.UUID{Bytes: ownerID, Valid: true})
	if err != nil {
		return nil, err
	}
	projects := make([]*project.Project, len(rows))
	for i, row := range rows {
		projects[i] = toProject(row)
	}
	return projects, nil
}

func (r *PostgresProjectRepository) Update(ctx context.Context, params project.UpdateProjectParams) (*project.Project, error) {
	row, err := r.q.UpdateProject(ctx, db.UpdateProjectParams{
		ID:          pgtype.UUID{Bytes: params.ID, Valid: true},
		OwnerID:     pgtype.UUID{Bytes: params.OwnerID, Valid: true},
		Name:        params.Name,
		Description: pgtype.Text{String: params.Description, Valid: params.Description != ""},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, project.ErrProjectNotFound
		}
		return nil, err
	}
	return toProject(row), nil
}

func toProject(row db.Project) *project.Project {
	p := &project.Project{
		ID:          uuid.UUID(row.ID.Bytes),
		OwnerID:     uuid.UUID(row.OwnerID.Bytes),
		Name:        row.Name,
		Slug:        row.Slug,
		Environment: row.Environment,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
	if row.Description.Valid {
		p.Description = row.Description.String
	}
	return p
}
