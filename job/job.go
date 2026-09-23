package job

import (
	"sonla/transcode/model"
	"sonla/transcode/util"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	for {
		var job model.Job
		if err := db.Where("status = ?", model.JOB_WAITING).Order("`created_at` ASC").Preload("App").First(&job).Error; err == nil {
			Transcode(db, job)
		} else if err == gorm.ErrRecordNotFound {
			time.Sleep(5 * time.Second)
		} else {
			util.Logger.Error("Job run", zap.String("error", err.Error()))
			time.Sleep(5 * time.Second)
		}
	}
}
