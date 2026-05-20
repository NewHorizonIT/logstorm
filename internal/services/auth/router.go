package auth

import "github.com/gin-gonic/gin"

type AuthRouter struct {
	handler *AuthHandler
}

func NewAuthRouter(handler *AuthHandler) *AuthRouter {
	return &AuthRouter{handler: handler}
}

func (ar *AuthRouter) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/login", ar.handler.LoginHandler)
	r.POST("/register", ar.handler.RegisterHandler)
	r.POST("/refresh", ar.handler.RefreshHandler)
	r.POST("/logout", ar.handler.LogoutHandler)
}
