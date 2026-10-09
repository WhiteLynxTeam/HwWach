package repository

import (
	"HwWach/internal/models"
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryWithStats struct {
	models.Category
	AssetsCount int64 `gorm:"column:assets_count" json:"assets_count"`
}

type CategoryRepo interface {
	Create(ctx context.Context, category *models.Category) error
	GetByUUID(ctx context.Context, id uuid.UUID) (*models.Category, error)
	GetByNormalizedName(ctx context.Context, normalizedName string, level int16, userUUID *uuid.UUID) (*models.Category, error)
	FindExactMatch(ctx context.Context, normalizedName string, userUUID uuid.UUID) (*models.Category, error)
	Search(ctx context.Context, query string, userUUID uuid.UUID, limit int) ([]*models.Category, error)
	Sync(ctx context.Context, userUUID uuid.UUID, since *time.Time) ([]*models.Category, []uuid.UUID, error)
	ListPending(ctx context.Context) ([]*models.Category, error)
	ListAdmin(ctx context.Context, level *int16, status *string, search string) ([]*CategoryWithStats, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementUsage(ctx context.Context, id uuid.UUID) error
}

type categoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) CategoryRepo {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) Create(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepo) GetByUUID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).Where("uuid = ?", id).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepo) GetByNormalizedName(ctx context.Context, normalizedName string, level int16, userUUID *uuid.UUID) (*models.Category, error) {
	var category models.Category
	query := r.db.WithContext(ctx).Where("normalized_name = ? AND level = ?", normalizedName, level)
	if level == 3 && userUUID != nil {
		query = query.Where("created_by = ?", *userUUID)
	}
	if err := query.First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

// FindExactMatch ищет точное совпадение по имени среди L1, L2 и собственных L3
func (r *categoryRepo) FindExactMatch(ctx context.Context, normalizedName string, userUUID uuid.UUID) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).
		Where("normalized_name = ? AND deleted_at IS NULL", normalizedName).
		Where("(level IN (1, 2) AND status = 'approved') OR (level = 3 AND created_by = ? AND status != 'rejected')", userUUID).
		Order("level ASC").
		First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// Search выполняет нечёткий (pg_trgm) и подстрочный поиск категорий
func (r *categoryRepo) Search(ctx context.Context, query string, userUUID uuid.UUID, limit int) ([]*models.Category, error) {
	if limit <= 0 {
		limit = 20
	}

	normQuery := strings.ToLower(strings.TrimSpace(query))
	var categories []*models.Category

	dbQuery := r.db.WithContext(ctx).Model(&models.Category{}).
		Where("deleted_at IS NULL").
		Where("(level IN (1, 2) AND status = 'approved') OR (level = 3 AND created_by = ? AND status != 'rejected')", userUUID)

	if normQuery == "" {
		// При пустом запросе отдаём популярные L1 и L2 категории
		err := dbQuery.
			Where("level IN (1, 2)").
			Order("level ASC, usage_count DESC, name ASC").
			Limit(limit).
			Find(&categories).Error
		return categories, err
	}

	likePattern := "%" + normQuery + "%"

	// Используем триграммы (%) и ILIKE подстроку
	err := dbQuery.
		Where("(normalized_name % ? OR normalized_name ILIKE ?)", normQuery, likePattern).
		Order(gorm.Expr("level ASC, similarity(normalized_name, ?) DESC, usage_count DESC, name ASC", normQuery)).
		Limit(limit).
		Find(&categories).Error

	return categories, err
}

// Sync возвращает список обновленных категорий и ID удаленных/объединенных с момента since
func (r *categoryRepo) Sync(ctx context.Context, userUUID uuid.UUID, since *time.Time) ([]*models.Category, []uuid.UUID, error) {
	var categories []*models.Category
	var deletedIDs []uuid.UUID

	if since == nil {
		// Полная синхронизация: все активные L1/L2 и одобренные/pending L3 пользователя
		err := r.db.WithContext(ctx).
			Where("deleted_at IS NULL").
			Where("(level IN (1, 2) AND status = 'approved') OR (level = 3 AND created_by = ? AND status != 'rejected')", userUUID).
			Order("level ASC, name ASC").
			Find(&categories).Error
		return categories, deletedIDs, err
	}

	// Инкрементальная синхронизация: всё что изменилось после since
	var allUpdated []*models.Category
	err := r.db.WithContext(ctx).Unscoped().
		Where("updated_at > ?", since).
		Where("(level IN (1, 2)) OR (level = 3 AND created_by = ?)", userUUID).
		Find(&allUpdated).Error
	if err != nil {
		return nil, nil, err
	}

	for _, c := range allUpdated {
		if c.DeletedAt.Valid || c.Status == models.CategoryStatusMerged || c.Status == models.CategoryStatusRejected {
			deletedIDs = append(deletedIDs, c.UUID)
		} else {
			categories = append(categories, c)
		}
	}

	return categories, deletedIDs, nil
}

// ListPending возвращает список L3 категорий на модерацию (для админа)
func (r *categoryRepo) ListPending(ctx context.Context) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.WithContext(ctx).
		Where("level = 3 AND status = ?", models.CategoryStatusPending).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepo) Update(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepo) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Category{}).
		Where("uuid = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

func (r *categoryRepo) ListAdmin(ctx context.Context, level *int16, status *string, search string) ([]*CategoryWithStats, error) {
	var results []*CategoryWithStats

	query := r.db.WithContext(ctx).
		Table("categories").
		Select("categories.*, COALESCE(COUNT(assets.uuid), 0) as assets_count").
		Joins("LEFT JOIN assets ON assets.category_uuid = categories.uuid AND assets.deleted_at IS NULL").
		Where("categories.deleted_at IS NULL")

	if level != nil {
		query = query.Where("categories.level = ?", *level)
	}
	if status != nil && *status != "" {
		query = query.Where("categories.status = ?", *status)
	}
	if search != "" {
		clean := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		query = query.Where("categories.normalized_name ILIKE ?", clean)
	}

	query = query.Group("categories.uuid").
		Order("categories.level ASC, categories.usage_count DESC, categories.name ASC")

	err := query.Find(&results).Error
	return results, err
}

func (r *categoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("uuid = ?", id).Delete(&models.Category{}).Error
}

