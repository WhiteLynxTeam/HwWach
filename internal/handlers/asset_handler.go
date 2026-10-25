package handlers

import (
	"HwWach/internal/authz"
	"HwWach/internal/dto"
	"HwWach/internal/middleware"
	"HwWach/internal/models"
	"HwWach/internal/services"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type assetHandler struct {
	assetSvc     services.AssetService
	lifecycleSvc services.LifecycleService
	photoSvc     services.PhotoService
	categorySvc  services.CategoryService
}

func NewAssetHandler(assetSvc services.AssetService, lifecycleSvc services.LifecycleService, photoSvc services.PhotoService, categorySvc services.CategoryService) AssetHandler {
	return &assetHandler{
		assetSvc:     assetSvc,
		lifecycleSvc: lifecycleSvc,
		photoSvc:     photoSvc,
		categorySvc:  categorySvc,
	}
}

func (a assetHandler) withPhotos(v *services.AssetView) dto.AssetWithPhotosResponse {
	return dto.AssetWithPhotosResponse{
		AssetResponse: viewToResponse(v),
		Photos:        photosToResponse(v.Asset.Photos, a.photoSvc.GetPublicURL),
	}
}

// CreateAsset godoc
// @Summary      Создать asset
// @Description  Карточка актива. МОЛ: не передан — создатель, если он МОЛ; "" — без МОЛ; другого человека назначает только админ
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateAssetRequest  true  "Данные asset"
// @Success      201      {object}  dto.AssetWithPhotosResponse
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Router       /assets [post]
// @Security     BearerAuth
func (a assetHandler) CreateAsset(c *gin.Context) {
	p, ok := middleware.RequirePerm(c, authz.PermAssetsPropose, authz.PermAssetsManage)
	if !ok {
		return
	}
	var req dto.CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	var catUUID *uuid.UUID
	if req.CategoryUUID != nil && *req.CategoryUUID != "" {
		parsed, err := uuid.Parse(*req.CategoryUUID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_uuid: " + err.Error()})
			return
		}
		catUUID = &parsed
	}

	categoryText := req.CategoryText
	if categoryText == nil || *categoryText == "" {
		if req.Category != "" {
			categoryText = &req.Category
		}
	}

	resolvedCat, err := a.categorySvc.ResolveOrCreateCategory(c.Request.Context(), p, catUUID, categoryText)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to resolve category: " + err.Error()})
		return
	}

	var inventoryNum *string
	if req.InventoryNum != nil {
		trimmed := strings.TrimSpace(*req.InventoryNum)
		if trimmed != "" {
			inventoryNum = &trimmed
		}
	}

	asset := &models.Asset{
		InventoryNum:     inventoryNum,
		Name:             req.Name,
		Category:         resolvedCat.Name,
		CategoryUUID:     &resolvedCat.UUID,
		Description:      req.Description,
		AccountingStatus: models.AccountingStatus(req.AccountingStatus),
		Condition:        models.Condition(req.Condition),
	}
	if asset.AccountingStatus != "" && !asset.AccountingStatus.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid accounting_status"})
		return
	}
	if asset.Condition != "" && !asset.Condition.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid condition"})
		return
	}

	if req.ClientID != nil {
		clientUUID, err := uuid.Parse(*req.ClientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id: " + err.Error()})
			return
		}
		asset.ClientID = &clientUUID
	}

	in := services.CreateAssetInput{Asset: asset}
	for _, idStr := range req.PhotoClientIDs {
		parsedUUID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid photo_client_id: " + err.Error()})
			return
		}
		in.PhotoClientIDs = append(in.PhotoClientIDs, parsedUUID)
	}
	if req.MolPersonID != nil {
		if *req.MolPersonID == "" {
			in.NoMol = true
		} else {
			id, err := uuid.Parse(*req.MolPersonID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mol_person_id"})
				return
			}
			in.MolPersonID = &id
		}
	}

	view, err := a.assetSvc.Create(c.Request.Context(), p, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, a.withPhotos(view))
}

func assetFilterFromQuery(c *gin.Context) services.AssetFilter {
	f := services.AssetFilter{
		AccountingStatus: c.Query("accounting_status"),
		Condition:        c.Query("condition"),
		Moderation:       c.Query("moderation_status"),
		NoMol:            c.Query("no_mol") == "true",
		Mine:             c.Query("mine") == "true",
		Overdue:          c.Query("overdue") == "true",
		Query:            c.Query("q"),
		IncludeInactive:  c.Query("include_inactive") == "true",
	}
	if s := c.Query("mol_person_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			f.MolPersonID = &id
		}
	}
	return f
}

// ListUserAssets godoc
// @Summary      Список assets
// @Description  С assets.view_all — все активы организации; иначе свои (МОЛ, работает с активом, завёл карточку). Списанные и утерянные скрыты, если не include_inactive=true
// @Tags         assets
// @Produce      json
// @Param        accounting_status  query  string  false  "in_use / in_stock / repair / lost / decommissioned"
// @Param        condition          query  string  false  "ok / needs_attention / faulty / unusable"
// @Param        mol_person_id      query  string  false  "Числится за человеком"
// @Param        no_mol             query  bool    false  "Без МОЛ"
// @Param        mine               query  bool    false  "Где у меня есть роль"
// @Param        overdue            query  bool    false  "Просрочен возврат из ремонта"
// @Param        q                  query  string  false  "Поиск по названию, инв. номеру, категории"
// @Success      200  {object}  dto.AssetListResponse
// @Failure      401  {object}  map[string]string
// @Router       /assets [get]
// @Security     BearerAuth
func (a assetHandler) ListUserAssets(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	views, _, err := a.assetSvc.List(c.Request.Context(), p, assetFilterFromQuery(c), 1, 0)
	if err != nil {
		writeErr(c, err)
		return
	}
	out := make([]dto.AssetWithPhotosResponse, 0, len(views))
	for _, v := range views {
		out = append(out, a.withPhotos(v))
	}
	c.JSON(http.StatusOK, dto.AssetListResponse{PersonID: p.PersonID.String(), Assets: out})
}

// ListUserAssetsPaginated godoc
// @Summary      Список assets с пагинацией и фотографиями
// @Description  Те же фильтры, что у GET /assets
// @Tags         assets
// @Produce      json
// @Param        page   query     int  false  "Номер страницы" default(1)
// @Param        limit  query     int  false  "Количество элементов на странице" default(10)
// @Success      200  {object}  dto.PaginatedAssetResponse
// @Failure      401  {object}  map[string]string
// @Router       /assets/paginated [get]
// @Security     BearerAuth
func (a assetHandler) ListUserAssetsPaginated(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	views, total, err := a.assetSvc.List(c.Request.Context(), p, assetFilterFromQuery(c), page, limit)
	if err != nil {
		writeErr(c, err)
		return
	}
	out := make([]dto.AssetWithPhotosResponse, 0, len(views))
	for _, v := range views {
		out = append(out, a.withPhotos(v))
	}
	c.JSON(http.StatusOK, dto.PaginatedAssetResponse{
		Assets: out,
		Total:  total,
		Page:   page,
		Limit:  limit,
		Pages:  int((total + int64(limit) - 1) / int64(limit)),
	})
}

// GetAsset godoc
// @Summary      Получить asset по ID
// @Description  Карточка с МОЛ, работающими с активом и открытым ремонтом
// @Tags         assets
// @Produce      json
// @Param        id  path      string  true  "ID asset (uuid)"
// @Success      200  {object}  dto.AssetWithPhotosResponse
// @Failure      404  {object}  map[string]string
// @Router       /assets/{id} [get]
// @Security     BearerAuth
func (a assetHandler) GetAsset(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	view, err := a.assetSvc.Get(c.Request.Context(), p, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, a.withPhotos(view))
}

// ListAssetPhotos godoc
// @Summary      Список фотографий asset
// @Tags         assets
// @Produce      json
// @Param        id  path      string  true  "ID asset (uuid)"
// @Success      200  {object}  map[string]interface{}
// @Router       /assets/{id}/photos [get]
// @Security     BearerAuth
func (a assetHandler) ListAssetPhotos(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	photos, err := a.assetSvc.ListPhotos(c.Request.Context(), p, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	plain := make([]models.Photo, 0, len(photos))
	for _, ph := range photos {
		plain = append(plain, *ph)
	}
	c.JSON(http.StatusOK, gin.H{
		"asset_uuid": id.String(),
		"photos":     photosToResponse(plain, a.photoSvc.GetPublicURL),
	})
}

// UpdateAsset godoc
// @Summary      Редактировать asset
// @Description  До одобрения — тот, кто завёл карточку; после — assets.manage. Статусы — через /status и ремонт
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        id       path      string                  true  "ID asset (uuid)"
// @Param        request  body      dto.UpdateAssetRequest  true  "Данные для обновления"
// @Success      200      {object}  dto.AssetWithPhotosResponse
// @Router       /assets/{id} [put]
// @Security     BearerAuth
func (a assetHandler) UpdateAsset(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if req.CategoryUUID != nil || req.CategoryText != nil || req.Category != nil {
		var catUUID *uuid.UUID
		if req.CategoryUUID != nil && *req.CategoryUUID != "" {
			parsed, err := uuid.Parse(*req.CategoryUUID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_uuid: " + err.Error()})
				return
			}
			catUUID = &parsed
		}
		categoryText := req.CategoryText
		if categoryText == nil || *categoryText == "" {
			if req.Category != nil && *req.Category != "" {
				categoryText = req.Category
			}
		}
		resolvedCat, err := a.categorySvc.ResolveOrCreateCategory(c.Request.Context(), p, catUUID, categoryText)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to resolve category: " + err.Error()})
			return
		}
		catName := resolvedCat.Name
		req.Category = &catName
		catUUIDStr := resolvedCat.UUID.String()
		req.CategoryUUID = &catUUIDStr
	}

	view, err := a.assetSvc.Update(c.Request.Context(), p, id, &req)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, a.withPhotos(view))
}

// DeleteAsset godoc
// @Summary      Удалить asset (assets.manage)
// @Tags         assets
// @Param        id  path  string  true  "ID asset (uuid)"
// @Success      200  {object}  map[string]string
// @Router       /assets/{id} [delete]
// @Security     BearerAuth
func (a assetHandler) DeleteAsset(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := a.assetSvc.Delete(c.Request.Context(), p, id); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ModerateAsset godoc
// @Summary      Одобрить / отклонить карточку (assets.manage)
// @Description  При одобрении можно сразу назначить МОЛ. Без МОЛ — актив попадёт в «Без МОЛ»
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        id       path  string                    true  "ID asset (uuid)"
// @Param        request  body  dto.ModerateAssetRequest  true  "Решение"
// @Success      200  {object}  dto.AssetWithPhotosResponse
// @Router       /assets/{id}/moderate [post]
// @Security     BearerAuth
func (a assetHandler) ModerateAsset(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.ModerateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var mol *uuid.UUID
	if req.MolPersonID != nil && *req.MolPersonID != "" {
		parsed, err := uuid.Parse(*req.MolPersonID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mol_person_id"})
			return
		}
		mol = &parsed
	}
	view, err := a.assetSvc.Moderate(c.Request.Context(), p, id, req.Status == "approved", req.AdminComment, mol)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, a.withPhotos(view))
}

// CheckInventoryUnique godoc
// @Summary      Проверить уникальность инвентарного номера в организации
// @Tags         assets
// @Produce      json
// @Param        num      query     string  true  "Инвентарный номер"
// @Param        exclude  query     string  false "Не считать этот актив (при редактировании)"
// @Success      200      {object}  map[string]bool
// @Router       /assets/check-inventory [get]
// @Security     BearerAuth
func (a assetHandler) CheckInventoryUnique(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	inventoryNum := strings.TrimSpace(c.Query("num"))
	if inventoryNum == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "num query parameter is required"})
		return
	}
	var exclude *uuid.UUID
	if s := c.Query("exclude"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			exclude = &id
		}
	}
	isUnique, err := a.assetSvc.IsInventoryNumUnique(c.Request.Context(), p.OrgID, inventoryNum, exclude)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"unique": isUnique})
}

// ChangeStatus godoc
// @Summary      Учётный статус и / или техническое состояние
// @Description  Админ или МОЛ актива; «требует внимания» и «неисправен» — любой, кто работает с активом. Списать — только админ (роли закрываются)
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        id       path  string                   true  "ID asset (uuid)"
// @Param        request  body  dto.ChangeStatusRequest  true  "Новые значения"
// @Success      200  {object}  dto.AssetWithPhotosResponse
// @Router       /assets/{id}/status [post]
// @Security     BearerAuth
func (a assetHandler) ChangeStatus(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	in := services.StatusInput{Comment: req.Comment}
	if req.AccountingStatus != nil {
		s := models.AccountingStatus(*req.AccountingStatus)
		in.AccountingStatus = &s
	}
	if req.Condition != nil {
		s := models.Condition(*req.Condition)
		in.Condition = &s
	}
	if err := a.lifecycleSvc.ChangeStatus(c.Request.Context(), p, id, in); err != nil {
		writeErr(c, err)
		return
	}
	a.GetAsset(c)
}

func parseOptTime(s *string) (*time.Time, bool) {
	if s == nil || *s == "" {
		return nil, true
	}
	if t, err := time.Parse(time.RFC3339, *s); err == nil {
		return &t, true
	}
	if t, err := time.Parse("2006-01-02", *s); err == nil {
		return &t, true
	}
	return nil, false
}

// StartRepair godoc
// @Summary      «В ремонт» (админ или МОЛ актива)
// @Description  Актив числится за тем же МОЛ. Отдали мастеру — обычная передача МОЛ → МОЛ, мастер сам ставит «В ремонт»
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        id       path  string                  true  "ID asset (uuid)"
// @Param        request  body  dto.RepairStartRequest  true  "Вид работ, ожидаемый возврат"
// @Success      200  {object}  dto.AssetWithPhotosResponse
// @Router       /assets/{id}/repair/start [post]
// @Security     BearerAuth
func (a assetHandler) StartRepair(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.RepairStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	expected, ok := parseOptTime(req.ExpectedReturnAt)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expected_return_at"})
		return
	}
	in := services.RepairStartInput{Kind: models.RepairKind(req.Kind), ExpectedReturnAt: expected, Comment: req.Comment}
	if req.RequestID != nil && *req.RequestID != "" {
		rid, err := uuid.Parse(*req.RequestID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request_id"})
			return
		}
		in.RequestID = &rid
	}
	if _, err := a.lifecycleSvc.StartRepair(c.Request.Context(), p, id, in); err != nil {
		writeErr(c, err)
		return
	}
	a.GetAsset(c)
}

// FinishRepair godoc
// @Summary      «Из ремонта» (текущий МОЛ или админ)
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        id       path  string                   true  "ID asset (uuid)"
// @Param        request  body  dto.RepairFinishRequest  true  "Итог ремонта"
// @Success      200  {object}  dto.AssetWithPhotosResponse
// @Router       /assets/{id}/repair/finish [post]
// @Security     BearerAuth
func (a assetHandler) FinishRepair(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.RepairFinishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	in := services.RepairFinishInput{ResultCondition: models.Condition(req.ResultCondition), WorkDone: req.WorkDone, Comment: req.Comment}
	if req.StatusAfter != nil && *req.StatusAfter != "" {
		s := models.AccountingStatus(*req.StatusAfter)
		in.StatusAfter = &s
	}
	if _, err := a.lifecycleSvc.FinishRepair(c.Request.Context(), p, id, in); err != nil {
		writeErr(c, err)
		return
	}
	a.GetAsset(c)
}

// History godoc
// @Summary      История актива: статусы и ремонты
// @Tags         assets
// @Produce      json
// @Param        id  path  string  true  "ID asset (uuid)"
// @Success      200  {object}  services.AssetHistory
// @Router       /assets/{id}/history [get]
// @Security     BearerAuth
func (a assetHandler) History(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	h, err := a.lifecycleSvc.History(c.Request.Context(), p, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	repairs := make([]*dto.RepairResponse, 0, len(h.Repairs))
	for i := range h.Repairs {
		repairs = append(repairs, repairToResponse(&h.Repairs[i]))
	}
	c.JSON(http.StatusOK, gin.H{"status": h.Status, "repairs": repairs})
}
