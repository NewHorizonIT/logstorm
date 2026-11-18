package user

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/response"
)

type UserHandler struct {
	userSvc *UserService
}

func NewUserHandler(userSvc *UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

type meResponse struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	Status     string `json:"status"`
	IsVerified bool   `json:"is_verified"`
	CreatedAt  string `json:"created_at"`
}

// authUserIDKey mirrors auth.UserIDKey — kept as a string literal to avoid an
// import cycle (auth imports user for UserProvider).
const authUserIDKey = "auth_user_id"

func userIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	id, exists := c.Get(authUserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	uid, ok := id.(uuid.UUID)
	return uid, ok
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "MISSING_TOKEN", "user not authenticated")
		return
	}

	u, err := h.userSvc.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(c, "USER_NOT_FOUND", "user not found")
			return
		}
		response.InternalServerError(c)
		return
	}

	response.OK(c, meResponse{
		ID:         u.ID.String(),
		Email:      u.Email,
		FullName:   u.FullName,
		Status:     u.Status,
		IsVerified: u.IsVerified,
		CreatedAt:  u.CreatedAt.Format(time.RFC3339),
	})
}
