package project

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/logger"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/response"
)

type ProjectHandler struct {
	svc *ProjectService
}

func NewProjectHandler(svc *ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

type createProjectRequest struct {
	Name        string `json:"name"        binding:"required"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Environment string `json:"environment"`
}

type updateProjectRequest struct {
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description"`
}

type projectResponse struct {
	ID          string `json:"id"`
	OwnerID     string `json:"owner_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Environment string `json:"environment"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (h *ProjectHandler) Create(c *gin.Context) {
	ownerID, ok := auth.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "MISSING_TOKEN", "authentication required")
		return
	}

	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	proj, err := h.svc.CreateProject(c.Request.Context(), ownerID, CreateProjectInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Environment: req.Environment,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, toProjectResponse(proj))
}

func (h *ProjectHandler) List(c *gin.Context) {
	ownerID, ok := auth.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "MISSING_TOKEN", "authentication required")
		return
	}

	projects, err := h.svc.ListProjects(c.Request.Context(), ownerID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	out := make([]projectResponse, len(projects))
	for i, p := range projects {
		out[i] = toProjectResponse(p)
	}
	response.OK(c, out)
}

func (h *ProjectHandler) Get(c *gin.Context) {
	ownerID, ok := auth.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "MISSING_TOKEN", "authentication required")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_PROJECT_ID", "project id must be a valid UUID")
		return
	}

	proj, err := h.svc.GetProject(c.Request.Context(), id, ownerID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, toProjectResponse(proj))
}

func (h *ProjectHandler) Update(c *gin.Context) {
	ownerID, ok := auth.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "MISSING_TOKEN", "authentication required")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_PROJECT_ID", "project id must be a valid UUID")
		return
	}

	var req updateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	proj, err := h.svc.UpdateProject(c.Request.Context(), ownerID, UpdateProjectInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, toProjectResponse(proj))
}

func (h *ProjectHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrProjectNotFound):
		response.NotFound(c, "PROJECT_NOT_FOUND", "project not found")
	case errors.Is(err, ErrSlugAlreadyExists):
		response.Conflict(c, "SLUG_ALREADY_EXISTS", "a project with this slug already exists")
	case errors.Is(err, ErrValidation):
		response.UnprocessableEntity(c, "VALIDATION_ERROR", err.Error())
	default:
		log := logger.FromContext(c.Request.Context())
		log.Error().Err(err).Msg("project: unexpected handler error")
		response.InternalServerError(c)
	}
}

func toProjectResponse(p *Project) projectResponse {
	return projectResponse{
		ID:          p.ID.String(),
		OwnerID:     p.OwnerID.String(),
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Environment: p.Environment,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}
