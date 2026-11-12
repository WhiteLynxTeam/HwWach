package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LocationKind уровень места — подсказка для экрана, глубина дерева не ограничена
type LocationKind string

const (
	LocationSettlement LocationKind = "settlement" // населённый пункт
	LocationSite       LocationKind = "site"       // объект / подразделение
	LocationRoom       LocationKind = "room"       // помещение
	LocationOther      LocationKind = "other"
)

func (k LocationKind) Valid() bool {
	switch k {
	case LocationSettlement, LocationSite, LocationRoom, LocationOther:
		return true
	}
	return false
}

// AssetLocationKind где находится актив
type AssetLocationKind string

const (
	AssetLocationUnset    AssetLocationKind = "unset"    // не указано
	AssetLocationPlace    AssetLocationKind = "place"    // в месте из справочника
	AssetLocationNone     AssetLocationKind = "none"     // места нет: почта, лицензия
	AssetLocationPortable AssetLocationKind = "portable" // носимое, у сотрудника: смартфон, ЭЦП
)

func (k AssetLocationKind) Valid() bool {
	switch k {
	case AssetLocationUnset, AssetLocationPlace, AssetLocationNone, AssetLocationPortable:
		return true
	}
	return false
}

// Location место организации. Дерево: parent_id, path — полный путь через « / »
type Location struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id" swaggertype:"string"`
	OrganizationID    uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id" swaggertype:"string"`
	ParentID          *uuid.UUID     `gorm:"type:uuid" json:"parent_id,omitempty" swaggertype:"string"`
	Name              string         `gorm:"type:varchar(255);not null" json:"name"`
	NormalizedName    string         `gorm:"type:varchar(255);not null" json:"-"`
	Kind              LocationKind   `gorm:"type:varchar(20);not null;default:'other'" json:"kind"`
	Path              string         `gorm:"type:text;not null" json:"path"`
	CreatedByPersonID *uuid.UUID     `gorm:"type:uuid" json:"created_by_person_id,omitempty" swaggertype:"string"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-" swaggertype:"string"`
}

// LocationPathSep разделитель уровней в path
const LocationPathSep = " / "
