package models

import (
	"time"

	"gorm.io/gorm"
)

type Request struct {
	ID        uint `gorm:"primaryKey"`
	DeviceID  uint `gorm:"not null;index"`
	UserID    uint `gorm:"not null;index"`
	CreatedAt time.Time
	Message   string `gorm:"type:text;not null"`
	PhotoURL  *string
	Status    string `gorm:"type:varchar(20);not null;default:'Оформлена'"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Device Device `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	User   User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
