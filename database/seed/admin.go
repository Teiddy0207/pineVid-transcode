package seed

import (
	"sonla/transcode/model"

	"gorm.io/gorm"
)

func Admins(db *gorm.DB) error {
	var admin model.Admin

	if err := db.Where("id > 0").Delete(&admin).Error; err != nil {
		return err
	}

	admin = model.Admin{
		ID:       1,
		Username: "admin",
		FullName: "Admin",
		Password: "12345678",
	}
	if err := admin.Create(db); err != nil {
		return err
	} else {
		return nil
	}
}
