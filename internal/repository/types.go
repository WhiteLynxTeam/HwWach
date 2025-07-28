package repository

import (
	"gorm.io/gorm"
)

type baseRepo struct {
	db *gorm.DB
}

type userRepo struct {
	baseRepo
}

type deviceRepo struct {
	baseRepo
}

type photoRepo struct {
	baseRepo
}
type requestRepo struct {
	baseRepo
}
