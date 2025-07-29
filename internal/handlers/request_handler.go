package handlers

import (
	"HwWach/internal/services"
	"github.com/gin-gonic/gin"
)

type requestHandler struct {
	reqSvc services.RequestService
}

func NewRequestHandler(reqSvc services.RequestService) RequestHandler {
	return &requestHandler{
		reqSvc: reqSvc,
	}
}

func (r requestHandler) CreateRequest(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (r requestHandler) GetRequest(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (r requestHandler) DeleteRequest(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
