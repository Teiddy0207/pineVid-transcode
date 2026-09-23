package model

import (
	"encoding/json"
	"time"
)

type Job struct {
	ID          string     `gorm:"column:id" json:"id" form:"-"`
	AppID       int64     `gorm:"app_id" json:"appID" form:"appID"`
	RequestID   string    `gorm:"request_id" json:"requestID" form:"requestID"`
	Source      string    `gorm:"source" json:"source" form:"source"`
	Destination string    `gorm:"destination" json:"destination" form:"destination"`
	Retry       int       `gorm:"retry" json:"retry" form:"retry"`
	HLS         bool      `gorm:"hls" json:"hls" form:"hls"`
	DASH        bool      `gorm:"dash" json:"dash" form:"dash"`
	Encrypt     bool      `gorm:"encrypt" json:"encrypt" form:"encrypt"`
	Key         string    `gorm:"key" json:"key" form:"key"`
	KeyUri      string    `gorm:"key_uri" json:"keyUri" form:"keyUri"`
	Callback    string    `gorm:"callback" json:"callback" form:"callback"`
	Status      string    `gorm:"status" json:"status" form:"status"`
	Error       string    `gorm:"error" json:"error" form:"error"`
	Progress    int       `gorm:"progress" json:"progress" form:"progress"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt" form:"-"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt" form:"-"`

	App *App `gorm:"foreignkey:AppID;association_foreignkey:ID" json:"app"`
}

type Jobs []Job

func (u *Job) TableName() string {
	return "jobs"
}

func (u Job) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

func (u Jobs) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}
