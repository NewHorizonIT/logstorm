package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK sends 200 with {"data": data}.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// Created sends 201 with {"data": data}.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

// NoContent sends 204 with no body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response with a consistent {"code", "message"} body.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorBody{Code: code, Message: message})
}

func BadRequest(c *gin.Context, code, message string) {
	Error(c, http.StatusBadRequest, code, message)
}

func Unauthorized(c *gin.Context, code, message string) {
	Error(c, http.StatusUnauthorized, code, message)
}

func Forbidden(c *gin.Context, code, message string) {
	Error(c, http.StatusForbidden, code, message)
}

func NotFound(c *gin.Context, code, message string) {
	Error(c, http.StatusNotFound, code, message)
}

func Conflict(c *gin.Context, code, message string) {
	Error(c, http.StatusConflict, code, message)
}

func UnprocessableEntity(c *gin.Context, code, message string) {
	Error(c, http.StatusUnprocessableEntity, code, message)
}

// InternalServerError sends 500 with a generic message — never leak internal details.
func InternalServerError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
}

// Abort is like Error but calls AbortWithStatusJSON — use inside middleware.
func Abort(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, errorBody{Code: code, Message: message})
}

func AbortUnauthorized(c *gin.Context, code, message string) {
	Abort(c, http.StatusUnauthorized, code, message)
}
