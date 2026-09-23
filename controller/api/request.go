package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-ozzo/ozzo-validation/v4"
	"net/url"
	"path"
	"strings"
)

func init() {
	validation.ErrRequired = validation.ErrRequired.SetMessage("Dữ liệu không được bỏ trống")
	validation.ErrLengthOutOfRange = validation.NewError("validation_length_out_of_range", "Dữ liệu dài từ {{.min}} đến {{.max}} ký tự")
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type ProfileRequest struct {
	FullName string `json:"full_name"`
}

type PasswordRequest struct {
	OldPassword     string `json:"oldPassword"`
	NewPassword     string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type AppRequest struct {
	Name   string `json:"name" form:"name"`
	Code   string `json:"code" form:"code"`
	Key    string `json:"key" form:"key"`
	Status string `json:"status" form:"status"`
}

type JobRequest struct {
	AppID       int64  `json:"appID" form:"appID"`
	RequestID   string `json:"requestID" form:"requestID"`
	Source      string `json:"source" form:"source"`
	Destination string `json:"destination" form:"destination"`
	HLS         bool   `json:"hls" form:"hls"`
	DASH        bool   `json:"dash" form:"dash"`
	Encrypt     bool   `json:"encrypt" form:"encrypt"`
	Key         string `json:"key" form:"key"`
	KeyUri      string `json:"keyUri" form:"keyUri"`
	Callback    string `json:"callback" form:"callback"`
	Status      string `json:"status" form:"status"`
}

type JobCreateRequest struct {
	Data struct {
		AppCode     string `json:"appCode" form:"appCode"`
		Callback    string `json:"callback" form:"callback"`
		DASH        bool   `json:"dash" form:"dash"`
		Destination string `json:"destination" form:"destination"`
		Encrypt     bool   `json:"encrypt" form:"encrypt"`
		HLS         bool   `json:"hls" form:"hls"`
		Key         string `json:"key" form:"key"`
		KeyUri      string `json:"keyUri" form:"keyUri"`
		RequestID   string `json:"requestID" form:"requestID"`
		Source      string `json:"source" form:"source"`
	} `json:"data" form:"data"`
	Checksum string `json:"checksum"`
}

func Host(c *gin.Context) string {
	var h = c.GetHeader("Host")
	var scheme = "http"

	if len(h) == 0 {
		h = c.Request.Host
	}

	if c.GetHeader("X-Forwarded-Proto") != "" {
		scheme = c.GetHeader("X-Forwarded-Proto")
	} else if c.Request.TLS != nil {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, h)
}

func JoinPath(base string, paths ...string) string {
	if u, err := url.Parse(base); err != nil {
		p := path.Join(paths...)
		return fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(p, "/"))
	} else {
		u.Path = path.Join(paths...)
		return u.String()
	}
}

func (request PasswordRequest) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.OldPassword, validation.Required, validation.Length(8, 200)),
		validation.Field(&request.NewPassword, validation.Required, validation.Length(8, 200)),
		validation.Field(&request.ConfirmPassword, validation.Required, validation.Length(8, 200)),
	)
}

func (request AppRequest) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.Name, validation.Required, validation.Length(1, 200)),
		validation.Field(&request.Code, validation.Required),
		validation.Field(&request.Key, validation.Required),
	)
}

func (request JobRequest) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.AppID, validation.Required),
		validation.Field(&request.RequestID, validation.Required),
		validation.Field(&request.Source, validation.Required),
		validation.Field(&request.Destination, validation.Required),
	)
}
