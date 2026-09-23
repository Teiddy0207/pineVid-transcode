package seed

import (
	"sonla/transcode/util"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	if err := Admins(db); err != nil {
		util.Logger.Error("Seed admin", zap.Error(err))
	}
}
