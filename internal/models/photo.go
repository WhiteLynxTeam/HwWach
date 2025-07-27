package models

import (
	"time"

	"gorm.io/gorm"
)

type Photo struct {
	ID        uint   `gorm:"primaryKey"`
	DeviceID  *uint  `gorm:"index"`
	UserID    uint   `gorm:"not null;index"`
	URL       string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Device *Device `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	User   User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
