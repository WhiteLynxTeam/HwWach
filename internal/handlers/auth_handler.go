package handlers

import (
	"HwWach/internal/services"
	"github.com/gin-gonic/gin"
)

type authHandler struct {
	authSvc services.AuthService
}

func NewAuthHandler(authSvc services.AuthService) AuthHandler {
	return &authHandler{
		authSvc: authSvc,
	}
}

func (a authHandler) Login(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a authHandler) Register(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a authHandler) Logout(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a authHandler) Refresh(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
