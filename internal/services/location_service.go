package services

import (
	"HwWach/internal/authz"
	"HwWach/internal/models"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LocationNode уровень места в запросе: от корня к листу. ID с телефона — чтобы место,
// заведённое без сети, получило тот же id на сервере. Есть ID в справочнике — имя не нужно
type LocationNode struct {
	ID   *uuid.UUID
	Name string
	Kind models.LocationKind
}

type LocationService interface {
	List(ctx context.Context, p *authz.Principal, query string) ([]models.Location, error)
	// Create заводит цепочку мест (или находит готовые) и возвращает последнее
	Create(ctx context.Context, p *authz.Principal, chain []LocationNode) (*models.Location, error)
}

type locationService struct {
	db *gorm.DB
}

func NewLocationService(db *gorm.DB) LocationService {
	return &locationService{db: db}
}

func (s *locationService) List(ctx context.Context, p *authz.Principal, query string) ([]models.Location, error) {
	q := s.db.WithContext(ctx).Where("organization_id = ?", p.OrgID)
	if qs := strings.TrimSpace(query); qs != "" {
		q = q.Where("lower(path) LIKE ?", "%"+strings.ToLower(qs)+"%")
	}
	var out []models.Location
	err := q.Order("path").Find(&out).Error
	return out, err
}

func (s *locationService) Create(ctx context.Context, p *authz.Principal, chain []LocationNode) (*models.Location, error) {
	if !p.Has(authz.PermAssetsPropose) && !p.Has(authz.PermAssetsManage) {
		return nil, errNoRights
	}
	var leaf *models.Location
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		leaf, err = resolveLocationChain(ctx, tx, p, chain)
		return err
	})
	return leaf, err
}

// normalizeLocationName «  Ферма   700 » и «ферма 700» — одно место
func normalizeLocationName(name string) string {
	s := strings.ToLower(strings.Join(strings.Fields(name), " "))
	return strings.ReplaceAll(s, "ё", "е")
}

func cleanLocationName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

// defaultLocationKind уровень по глубине: населённый пункт → объект → помещение
func defaultLocationKind(depth int) models.LocationKind {
	switch depth {
	case 0:
		return models.LocationSettlement
	case 1:
		return models.LocationSite
	case 2:
		return models.LocationRoom
	}
	return models.LocationOther
}

// resolveLocationChain проходит цепочку от корня: узел ищется по id, затем по имени под тем же
// родителем, иначе заводится. Так два телефона, без сети заведшие «Ферма 700», получат одно место
func resolveLocationChain(ctx context.Context, tx *gorm.DB, p *authz.Principal, chain []LocationNode) (*models.Location, error) {
	if len(chain) == 0 {
		return nil, BadRequest("Не указано место")
	}
	if len(chain) > 10 {
		return nil, BadRequest("Слишком глубокая вложенность места")
	}
	var parent *models.Location
	for depth, n := range chain {
		loc, err := resolveLocationNode(ctx, tx, p, parent, n, depth)
		if err != nil {
			return nil, err
		}
		parent = loc
	}
	return parent, nil
}

func resolveLocationNode(ctx context.Context, tx *gorm.DB, p *authz.Principal, parent *models.Location, n LocationNode, depth int) (*models.Location, error) {
	db := tx.WithContext(ctx)
	var loc models.Location
	if n.ID != nil {
		err := db.Where("id = ? AND organization_id = ?", *n.ID, p.OrgID).First(&loc).Error
		if err == nil {
			return &loc, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	name := cleanLocationName(n.Name)
	if name == "" {
		return nil, BadRequest("Место не найдено")
	}
	if len([]rune(name)) > 255 {
		return nil, BadRequest("Слишком длинное название места")
	}
	norm := normalizeLocationName(name)
	var parentID *uuid.UUID
	if parent != nil {
		parentID = &parent.ID
	}

	q := db.Where("organization_id = ? AND normalized_name = ?", p.OrgID, norm)
	if parentID == nil {
		q = q.Where("parent_id IS NULL")
	} else {
		q = q.Where("parent_id = ?", *parentID)
	}
	err := q.First(&loc).Error
	if err == nil {
		return &loc, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	kind := n.Kind
	if !kind.Valid() {
		kind = defaultLocationKind(depth)
	}
	path := name
	if parent != nil {
		path = parent.Path + models.LocationPathSep + name
	}
	// id с телефона берём, если он свободен (удалённое место или чужая организация его не держат)
	id := newID()
	if n.ID != nil {
		var cnt int64
		if err := db.Unscoped().Model(&models.Location{}).Where("id = ?", *n.ID).Count(&cnt).Error; err != nil {
			return nil, err
		}
		if cnt == 0 {
			id = *n.ID
		}
	}
	personID := p.PersonID
	loc = models.Location{
		ID:                id,
		OrganizationID:    p.OrgID,
		ParentID:          parentID,
		Name:              name,
		NormalizedName:    norm,
		Kind:              kind,
		Path:              path,
		CreatedByPersonID: &personID,
	}
	if err := db.Create(&loc).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, Conflict("Такое место только что добавил кто-то другой, повторите")
		}
		return nil, err
	}
	return &loc, nil
}

// applyAssetLocation ставит активу место. kind пустой: есть цепочка — place, нет — не указано
func applyAssetLocation(ctx context.Context, tx *gorm.DB, p *authz.Principal, a *models.Asset, kind models.AssetLocationKind, chain []LocationNode) error {
	if kind == "" {
		if len(chain) > 0 {
			kind = models.AssetLocationPlace
		} else {
			kind = models.AssetLocationUnset
		}
	}
	if !kind.Valid() {
		return BadRequest("invalid location_kind")
	}
	a.LocationKind = kind
	a.Location = nil
	if kind != models.AssetLocationPlace {
		a.LocationID = nil
		return nil
	}
	leaf, err := resolveLocationChain(ctx, tx, p, chain)
	if err != nil {
		return err
	}
	a.LocationID = &leaf.ID
	return nil
}

// locationSubtreeSQL id места и всех вложенных — для фильтра «всё в Каленковичах»
const locationSubtreeSQL = `WITH RECURSIVE sub AS (
	SELECT id FROM locations WHERE id = ? AND organization_id = ?
	UNION ALL
	SELECT l.id FROM locations l JOIN sub ON l.parent_id = sub.id
) SELECT id FROM sub`
