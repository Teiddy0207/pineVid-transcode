package model

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type Admin struct {
	ID           int64     `gorm:"column:id" json:"id"`
	FullName     string    `gorm:"column:full_name" json:"fullName"`
	Username     string    `gorm:"column:username" json:"username"`
	Password     string    `gorm:"-" json:"password"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Admin) TableName() string {
	return "admins"
}

func (u *Admin) Create(db *gorm.DB) error {
	ph, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithStack(err)
	}
	u.PasswordHash = string(ph)
	return db.Create(u).Error
}
