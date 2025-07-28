package handlers

import (
	"HwWach/internal/services"
	"github.com/gin-gonic/gin"
)

type userHandler struct {
	userSvc services.UserService
}

func NewUserHandler(userSvc services.UserService) UserHandler {
	return &userHandler{
		userSvc: userSvc,
	}
}

func (u userHandler) GetProfile(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (u userHandler) UpdateProfile(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (u userHandler) ChangePassword(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (u userHandler) ListDevices(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
