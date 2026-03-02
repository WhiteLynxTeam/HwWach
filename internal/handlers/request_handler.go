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

// CreateRequest godoc
// @Summary      Создать запрос
// @Description  Создание нового запроса в системе
// @Tags         requests
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Данные запроса"
// @Success      201  {object}  models.Request
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /requests [post]
// @Security     BearerAuth
func (r requestHandler) CreateRequest(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

// GetRequest godoc
// @Summary      Получить запрос по ID
// @Description  Получение информации о запросе по идентификатору
// @Tags         requests
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "ID запроса (uuid)"
// @Success      200  {object}  models.Request
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /requests/{id} [get]
// @Security     BearerAuth
func (r requestHandler) GetRequest(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

// DeleteRequest godoc
// @Summary      Удалить запрос
// @Description  Удаление запроса по идентификатору
// @Tags         requests
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "ID запроса (uuid)"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /requests/{id} [delete]
// @Security     BearerAuth
func (r requestHandler) DeleteRequest(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
