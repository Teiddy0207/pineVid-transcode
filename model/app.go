package model

import (
	"encoding/json"
	"time"
)

type App struct {
	ID        int64     `gorm:"column:id" json:"id" form:"-"`
	Name      string    `gorm:"column:name" json:"name" form:"name"`
	Code      string    `gorm:"column:code" json:"code" form:"code"`
	Key       string    `gorm:"column:key" json:"key" form:"key"`
	Status    string    `gorm:"column:status" json:"status" form:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt" form:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt" form:"-"`
}

type Apps []App

func (u *App) TableName() string {
	return "apps"
}

func (u App) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

func (u Apps) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}
