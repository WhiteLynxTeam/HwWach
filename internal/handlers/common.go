package handlers

import (
	"HwWach/internal/dto"
	"HwWach/internal/models"
	"HwWach/internal/services"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// writeErr ошибки сервисов с понятным сообщением и статусом; прочие — 500
func writeErr(c *gin.Context, err error) {
	var appErr *services.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Status, gin.H{"error": appErr.Msg})
		return
	}
	c.Error(err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid format"})
		return uuid.Nil, false
	}
	return id, true
}

func fmtTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func uuidStr(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}

func roleToResponse(r *models.AssetRole) dto.AssetRoleResponse {
	return dto.AssetRoleResponse{
		ID:                r.ID.String(),
		PersonID:          r.PersonID.String(),
		Role:              string(r.Role),
		StartedAt:         r.StartedAt.Format(time.RFC3339),
		EndedAt:           fmtTime(r.EndedAt),
		EndReason:         r.EndReason,
		AckStatus:         string(r.AckStatus),
		AckReason:         r.AckReason,
		AckAt:             fmtTime(r.AckAt),
		DisputeResolvedAt: fmtTime(r.DisputeResolvedAt),
		AssignedByPerson:  uuidStr(r.AssignedByPersonID),
	}
}

func repairToResponse(r *models.AssetRepair) *dto.RepairResponse {
	if r == nil {
		return nil
	}
	resp := &dto.RepairResponse{
		ID:               r.ID.String(),
		Kind:             string(r.Kind),
		StartedAt:        r.StartedAt.Format(time.RFC3339),
		ExpectedReturnAt: fmtTime(r.ExpectedReturnAt),
		ReturnedAt:       fmtTime(r.ReturnedAt),
		// просрочен ожидаемый возврат — пометка в карточке и в списке (3.6)
		Overdue:         r.ReturnedAt == nil && r.ExpectedReturnAt != nil && r.ExpectedReturnAt.Before(time.Now()),
		StatusBefore:    string(r.StatusBefore),
		WorkDone:        r.WorkDone,
		Comment:         r.Comment,
		RequestID:       uuidStr(r.RequestID),
		CreatedByPerson: r.CreatedByPersonID.String(),
	}
	if r.ResultCondition != nil {
		s := string(*r.ResultCondition)
		resp.ResultCondition = &s
	}
	return resp
}

// assetToResponse конвертирует модель Asset в DTO AssetResponse
func assetToResponse(asset *models.Asset) dto.AssetResponse {
	resp := dto.AssetResponse{
		UUID:             asset.UUID.String(),
		OrganizationID:   asset.OrganizationID.String(),
		InventoryNum:     asset.InventoryNum,
		Name:             asset.Name,
		Category:         asset.Category,
		Description:      asset.Description,
		AccountingStatus: string(asset.AccountingStatus),
		Condition:        string(asset.Condition),
		ModerationStatus: string(asset.ModerationStatus),
		CreatedByAccount: asset.CreatedByAccountID.String(),
		CreatedByPerson:  asset.CreatedByPersonID.String(),
		Users:            []dto.AssetRoleResponse{},
		AdminComment:     asset.AdminComment,
		CreatedAt:        asset.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        asset.UpdatedAt.Format(time.RFC3339),
		VerifiedAt:       fmtTime(asset.VerifiedAt),
		CategoryUUID:     uuidStr(asset.CategoryUUID),
		ClientID:         uuidStr(asset.ClientID),
	}
	return resp
}

func viewToResponse(v *services.AssetView) dto.AssetResponse {
	resp := assetToResponse(v.Asset)
	if v.Mol != nil {
		r := roleToResponse(v.Mol)
		resp.Mol = &r
	}
	for i := range v.Users {
		resp.Users = append(resp.Users, roleToResponse(&v.Users[i]))
	}
	resp.OpenRepair = repairToResponse(v.OpenRepair)
	return resp
}

func photosToResponse(photos []models.Photo, publicURL func(string) string) []dto.PhotoResponse {
	out := make([]dto.PhotoResponse, 0, len(photos))
	for _, photo := range photos {
		pr := dto.PhotoResponse{
			UUID:      photo.UUID.String(),
			URL:       publicURL(photo.URL),
			CreatedAt: photo.CreatedAt.Format(time.RFC3339),
			ClientID:  uuidStr(photo.ClientID),
		}
		out = append(out, pr)
	}
	return out
}
