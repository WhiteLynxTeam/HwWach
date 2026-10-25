package services

import (
	"HwWach/internal/authz"
	"HwWach/internal/dto"
	"HwWach/internal/models"
	"HwWach/internal/repository"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type AssetChangeRequestService interface {
	CreateRequest(ctx context.Context, p *authz.Principal, assetUUID uuid.UUID, req dto.CreateChangeRequestInput) (*models.AssetChangeRequest, error)
	ApproveRequest(ctx context.Context, p *authz.Principal, reqUUID uuid.UUID, req dto.ApproveChangeRequestInput) error
	ListPending(ctx context.Context, p *authz.Principal) ([]*models.AssetChangeRequest, error)
}

type assetChangeRequestService struct {
	repo   repository.AssetChangeRequestRepo
	assets AssetService
	db     *assetService
}

func NewAssetChangeRequestService(repo repository.AssetChangeRequestRepo, assets AssetService) AssetChangeRequestService {
	return &assetChangeRequestService{repo: repo, assets: assets, db: assets.(*assetService)}
}

// CreateRequest заявку на изменение или удаление подаёт тот, кто завёл карточку, или МОЛ актива
func (s *assetChangeRequestService) CreateRequest(ctx context.Context, p *authz.Principal, assetUUID uuid.UUID, req dto.CreateChangeRequestInput) (*models.AssetChangeRequest, error) {
	acc, err := loadAsset(ctx, s.db.db, p, assetUUID, false)
	if err != nil {
		return nil, err
	}
	if acc.Asset.CreatedByPersonID != p.PersonID && !acc.IsMol {
		return nil, Forbidden("you don't have permission to modify this asset")
	}

	existingReq, err := s.repo.GetPendingByAssetID(ctx, assetUUID)
	if err != nil {
		return nil, err
	}
	if existingReq != nil {
		return nil, Conflict("asset already has a pending change request")
	}

	var clientUUID *uuid.UUID
	if req.ClientID != nil {
		parsed, err := uuid.Parse(*req.ClientID)
		if err != nil {
			return nil, BadRequest("invalid client_id")
		}
		clientUUID = &parsed
	}

	newReq := &models.AssetChangeRequest{
		OrganizationID: p.OrgID,
		ClientID:       clientUUID,
		AssetUUID:      assetUUID,
		AccountUUID:    p.AccountID,
		PersonUUID:     p.PersonID,
		RequestType:    models.RequestType(req.Type),
		ProposedData:   []byte(req.ProposedData),
		Reason:         req.Reason,
		Status:         models.ModerationPending,
	}

	if err := s.repo.Create(ctx, newReq); err != nil {
		return nil, err
	}

	return newReq, nil
}

// ApproveRequest право change_requests.review (раньше проверки не было — любой вошедший одобрял чужие заявки)
func (s *assetChangeRequestService) ApproveRequest(ctx context.Context, p *authz.Principal, reqUUID uuid.UUID, req dto.ApproveChangeRequestInput) error {
	if !p.Has(authz.PermChangeRequestsReview) {
		return errNoRights
	}
	changeReq, err := s.repo.GetByUUID(ctx, p.OrgID, reqUUID)
	if err != nil {
		return NotFound("request not found")
	}

	if changeReq.Status != models.ModerationPending {
		return Conflict("request is not in pending status")
	}

	changeReq.Status = models.ModerationStatus(req.Status)
	changeReq.AdminComment = req.AdminComment
	changeReq.ProcessedByAccountID = &p.AccountID
	changeReq.ProcessedByPersonID = &p.PersonID

	if changeReq.Status == models.ModerationApproved {
		switch changeReq.RequestType {
		case models.RequestTypeUpdate:
			var updateData dto.UpdateAssetRequest
			if len(changeReq.ProposedData) > 0 {
				if err := json.Unmarshal(changeReq.ProposedData, &updateData); err != nil {
					return BadRequest("failed to parse proposed data: " + err.Error())
				}
			}
			// применяем от имени проверяющего: у него assets.manage или review
			reviewer := *p
			reviewer.Perms = map[string]bool{authz.PermAssetsManage: true, authz.PermAssetsViewAll: true}
			if _, err := s.assets.Update(ctx, &reviewer, changeReq.AssetUUID, &updateData); err != nil {
				var appErr *AppError
				if errors.As(err, &appErr) {
					return err
				}
				return errors.New("failed to update asset")
			}
		case models.RequestTypeDelete:
			reviewer := *p
			reviewer.Perms = map[string]bool{authz.PermAssetsManage: true, authz.PermAssetsViewAll: true}
			if err := s.assets.Delete(ctx, &reviewer, changeReq.AssetUUID); err != nil {
				return err
			}
		}
	}

	return s.repo.UpdateStatus(ctx, changeReq)
}

func (s *assetChangeRequestService) ListPending(ctx context.Context, p *authz.Principal) ([]*models.AssetChangeRequest, error) {
	if !p.Has(authz.PermChangeRequestsReview) {
		return nil, errNoRights
	}
	return s.repo.ListPending(ctx, p.OrgID)
}
