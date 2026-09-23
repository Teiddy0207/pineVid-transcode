package api

import (
	"gorm.io/gorm"
)

type base struct {
	DB    *gorm.DB
}

func NewController(DB *gorm.DB) *base {
	return &base{
		DB:    DB,
	}
}
