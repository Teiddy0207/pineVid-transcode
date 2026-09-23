package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sonla/transcode/model"
	"sonla/transcode/util"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (b *base) GetJobs(c *gin.Context) {
	var total int64
	var limit int
	var page int
	var data = make(model.Jobs, 0)
	var search = c.Query("search")
	var status = c.Query("status")
	var appID = c.Query("appID")
	var draw, _ = strconv.ParseInt(c.Query("draw"), 10, 64)

	if v, err := strconv.ParseInt(c.Query("limit"), 10, 32); err != nil {
		limit = 20
	} else {
		limit = int(v)
	}
	if v, err := strconv.ParseInt(c.Query("page"), 10, 32); err != nil {
		page = 0
	} else {
		page = int(v)
	}

	/* Query database */
	qr := b.DB.Model(model.Job{})
	if len(search) > 0 {
		qr = qr.Or("jobs.request_id LIKE ?", fmt.Sprintf("%s%s%s", "%", search, "%"))
		qr = qr.Or("jobs.source LIKE ?", fmt.Sprintf("%s%s%s", "%", search, "%"))
		qr = qr.Or("jobs.destination LIKE ?", fmt.Sprintf("%s%s%s", "%", search, "%"))
	}
	if len(status) > 0 {
		qr = qr.Where("jobs.status = ?", status)
	}
	if len(appID) > 0 {
		qr = qr.Where("jobs.app_id = ?", appID)
	}

	qr = qr.Count(&total).Order("id ASC").Preload("App")

	if page >= 0 {
		qr = qr.Limit(limit).Offset(page * limit)
	}

	if err := qr.Find(&data).Error; err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{
			"draw":            draw,
			"data":            data,
			"recordsTotal":    total,
			"recordsFiltered": total,
			"error":           "Lỗi truy vấn cơ sở dữ liệu",
		})
	} else {
		c.JSON(http.StatusOK, map[string]interface{}{
			"draw":            draw,
			"data":            data,
			"recordsTotal":    total,
			"recordsFiltered": total,
		})
	}
}

func (b *base) GetJob(c *gin.Context) {
	var job model.Job

	var qr = b.DB.Model(job).Where("`id` = ?", c.Param("id")).Preload("App")
	if err := qr.First(&job).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": "Lỗi truy vấn cơ sở dữ liệu",
			"error":   err.Error(),
		})
	} else if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"code":    http.StatusNotFound,
			"message": "Không tìm thấy thông tin dữ liệu",
		})
	} else {
		c.JSON(http.StatusOK, job)
	}
}

func (b *base) PostJob(c *gin.Context) {
	var request JobRequest
	var job model.Job

	/* Bind request */
	if err := c.MustBindWith(&request, binding.JSON); err != nil {
		if err := c.MustBindWith(&request, binding.FormMultipart); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "Dữ liệu không hợp lệ",
			})
			return
		}
	}

	/* Validation */
	if err := request.Validate(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"code":    http.StatusUnprocessableEntity,
			"message": "Dữ liệu không hợp lệ",
			"errors":  err,
		})
		return
	} else {
		job = model.Job{
			ID:          uuid.New().String(),
			AppID:       request.AppID,
			RequestID:   request.RequestID,
			Source:      request.Source,
			Destination: request.Destination,
			Retry:       0,
			HLS:         request.HLS,
			DASH:        request.DASH,
			Encrypt:     request.Encrypt,
			Key:         request.Key,
			KeyUri:      request.KeyUri,
			Callback:    request.Callback,
			Status:      model.JOB_WAITING,
			Progress:    0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	/* Create job */
	if err := b.DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Lỗi lưu dữ liệu",
		})
		return
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"code":    http.StatusCreated,
			"message": "Thêm mới dữ liệu thành công",
		})
	}
}

func (b *base) PutJob(c *gin.Context) {
	var request JobRequest
	var job model.Job

	/* Bind request */
	if err := c.MustBindWith(&request, binding.JSON); err != nil {
		if err := c.MustBindWith(&request, binding.FormMultipart); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "Dữ liệu không hợp lệ",
			})
			return
		}
	}

	/* Find job */
	var qr = b.DB.Model(job).Where("`id` = ?", c.Param("id"))
	if err := qr.First(&job).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": "Lỗi truy vấn cơ sở dữ liệu",
			"error":   err.Error(),
		})
		return
	} else if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"code":    http.StatusNotFound,
			"message": "Không tìm thấy thông tin dữ liệu",
		})
		return
	}

	/* Validation */
	if err := request.Validate(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"code":    http.StatusUnprocessableEntity,
			"message": "Dữ liệu không hợp lệ",
			"errors":  err,
		})
		return
	} else {
		job.AppID = request.AppID
		job.RequestID = request.RequestID
		job.Source = request.Source
		job.Destination = request.Destination
		job.Retry = 0
		job.HLS = request.HLS
		job.DASH = request.DASH
		job.Encrypt = request.Encrypt
		job.Key = request.Key
		job.KeyUri = request.KeyUri
		job.Callback = request.Callback
		job.Status = request.Status
		job.UpdatedAt = time.Now()

		if request.Status == model.JOB_WAITING {
			job.Progress = 0
		}
		if request.Status == model.JOB_DONE {
			job.Progress = 100
		}
	}

	/* Update job */
	if err := b.DB.Save(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Lỗi lưu dữ liệu",
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "Cập nhật thông tin thành công",
		})
	}
}

func (b *base) DeleteJob(c *gin.Context) {
	var job model.Job

	/* Find job */
	var qr = b.DB.Model(job).Where("`id` = ?", c.Param("id"))
	if err := qr.First(&job).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": "Lỗi truy vấn cơ sở dữ liệu",
			"error":   err.Error(),
		})
		return
	} else if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"code":    http.StatusNotFound,
			"message": "Không tìm thấy thông tin dữ liệu",
		})
		return
	}

	/* Update job */
	if err := b.DB.Delete(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Lỗi lưu dữ liệu",
		})
		return
	} else {

		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "Xóa thông tin thành công",
		})
	}
}

func (b *base) PostJobCreate(c *gin.Context) {
	var request JobCreateRequest
	var job model.Job
	var app model.App

	/* Bind request */
	if err := c.MustBindWith(&request, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Dữ liệu không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	/* Find app */
	if err := b.DB.Where("code = ?", request.Data.AppCode).First(&app).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": "Lỗi truy vấn cơ sở dữ liệu",
			"error":   err.Error(),
		})
		return
	} else if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusForbidden, map[string]interface{}{
			"code":    http.StatusForbidden,
			"message": "Ứng dụng không tồn tại!",
		})
		return
	} else if app.Status != "active" {
		c.JSON(http.StatusForbidden, map[string]interface{}{
			"code":    http.StatusForbidden,
			"message": "Ứng dụng không được kích hoạt!",
		})
		return
	}

	/* Validate checksum */
	dataJson, _ := json.Marshal(request.Data)
	if util.MD5Hash(string(dataJson)+app.Key) != strings.ToUpper(request.Checksum) {
		c.JSON(http.StatusForbidden, map[string]interface{}{
			"code":    http.StatusForbidden,
			"message": "Checksum không hợp lệ!",
		})
		return
	}

	/* Find job */
	if err := b.DB.Where("`request_id` = ? AND `app_id` = ?", request.Data.RequestID, app.ID).First(&job).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": "Lỗi truy vấn cơ sở dữ liệu",
			"error":   err.Error(),
		})
		return
	} else if err == gorm.ErrRecordNotFound {
		job = model.Job{
			ID:          uuid.NewString(),
			AppID:       app.ID,
			RequestID:   request.Data.RequestID,
			Source:      request.Data.Source,
			Destination: request.Data.Destination,
			Retry:       0,
			HLS:         request.Data.HLS,
			DASH:        request.Data.DASH,
			Encrypt:     request.Data.Encrypt,
			Key:         request.Data.Key,
			KeyUri:      request.Data.KeyUri,
			Callback:    request.Data.Callback,
			Status:      model.JOB_WAITING,
			Progress:    0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := b.DB.Create(&job).Error; err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"code":    http.StatusInternalServerError,
				"message": "Lỗi truy vấn cơ sở dữ liệu",
				"error":   err.Error(),
			})
			return
		} else {
			c.JSON(http.StatusCreated, map[string]interface{}{
				"code":    http.StatusCreated,
				"message": "Thêm yêu cầu thành công!",
			})
			return
		}
	} else {
		job.Source = request.Data.Source
		job.Destination = request.Data.Destination
		job.HLS = request.Data.HLS
		job.DASH = request.Data.DASH
		job.Encrypt = request.Data.Encrypt
		job.Key = request.Data.Key
		job.KeyUri = request.Data.KeyUri
		job.Callback = request.Data.Callback
		job.Retry = 0
		job.Progress = 0
		job.Status = model.JOB_WAITING
		job.UpdatedAt = time.Now().UTC()
		if err := b.DB.Save(&job).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Lỗi lưu dữ liệu",
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"message": "Cập nhật thông tin thành công",
			})
		}
	}
}

func (b *base) PostJobRetry(c *gin.Context) {
	var request JobCreateRequest
	var job model.Job
	var app model.App

	/* Bind request */
	if err := c.MustBindWith(&request, binding.JSON); err != nil {
		if err := c.MustBindWith(&request, binding.FormMultipart); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "Dữ liệu không hợp lệ",
			})
			return
		}
	}

	/* Find app */
	if err := b.DB.Where("code = ?", request.Data.AppCode).First(&app).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": "Lỗi truy vấn cơ sở dữ liệu",
			"error":   err.Error(),
		})
		return
	} else if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusForbidden, map[string]interface{}{
			"code":    http.StatusForbidden,
			"message": "Ứng dụng không tồn tại!",
		})
		return
	} else if app.Status != "active" {
		c.JSON(http.StatusForbidden, map[string]interface{}{
			"code":    http.StatusForbidden,
			"message": "Ứng dụng không được kích hoạt!",
		})
		return
	}

	/* Validate checksum */
	dataJson, _ := json.Marshal(request.Data)
	if util.MD5Hash(string(dataJson)+app.Key) != strings.ToUpper(request.Checksum) {
		c.JSON(http.StatusForbidden, map[string]interface{}{
			"code":    http.StatusForbidden,
			"message": "Checksum không hợp lệ!",
		})
		return
	}

	/* Find job */
	if err := b.DB.Where("`request_id` = ? AND `app_id` = ?", request.Data.RequestID, app.ID).First(&job).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": "Lỗi truy vấn cơ sở dữ liệu",
			"error":   err.Error(),
		})
		return
	} else if err == gorm.ErrRecordNotFound {
		job = model.Job{
			ID:          uuid.NewString(),
			AppID:       app.ID,
			RequestID:   request.Data.RequestID,
			Source:      request.Data.Source,
			Destination: request.Data.Destination,
			Retry:       0,
			HLS:         request.Data.HLS,
			DASH:        request.Data.DASH,
			Encrypt:     request.Data.Encrypt,
			Key:         request.Data.Key,
			KeyUri:      request.Data.KeyUri,
			Callback:    request.Data.Callback,
			Status:      model.JOB_WAITING,
			Progress:    0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := b.DB.Create(&job).Error; err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"code":    http.StatusInternalServerError,
				"message": "Lỗi truy vấn cơ sở dữ liệu",
				"error":   err.Error(),
			})
			return
		} else {
			c.JSON(http.StatusCreated, map[string]interface{}{
				"code":    http.StatusCreated,
				"message": "Thêm yêu cầu thành công!",
			})
			return
		}
	} else {
		if job.Status == model.JOB_RUNNING {
			c.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    http.StatusForbidden,
				"message": "Yêu cầu đang thực thi!",
			})
			return
		}
		job.Source = request.Data.Source
		job.Destination = request.Data.Destination
		job.HLS = request.Data.HLS
		job.DASH = request.Data.DASH
		job.Encrypt = request.Data.Encrypt
		job.Key = request.Data.Key
		job.KeyUri = request.Data.KeyUri
		job.Callback = request.Data.Callback
		job.Retry = 0
		job.Progress = 0
		job.Status = model.JOB_WAITING
		job.UpdatedAt = time.Now().UTC()
		if err := b.DB.Save(&job).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Lỗi lưu dữ liệu",
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"message": "Cập nhật thông tin thành công",
			})
		}
	}
}
