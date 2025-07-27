package models

import (
	"time"

	"gorm.io/gorm"
)

type Device struct {
	ID            uint   `gorm:"primaryKey"`
	InventoryNum  string `gorm:"unique;not null"`
	Type          string `gorm:"not null"`
	Specification string `gorm:"type:text"`
	Status        string `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	Photos   []Photo   `gorm:"foreignKey:DeviceID"`
	Requests []Request `gorm:"foreignKey:DeviceID"`
}
