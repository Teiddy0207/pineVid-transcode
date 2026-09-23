package api

import (
	"fmt"
	"net/http"
	"sonla/transcode/model"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

func (b *base) GetApps(c *gin.Context) {
	var total int64
	var limit int
	var page int
	var data = make(model.Apps, 0)
	var search = c.Query("search")
	var status = c.Query("status")
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
	qr := b.DB.Model(model.App{})
	if len(search) > 0 {
		qr = qr.Or("apps.name LIKE ?", fmt.Sprintf("%s%s%s", "%", search, "%"))
		qr = qr.Or("apps.code LIKE ?", fmt.Sprintf("%s%s%s", "%", search, "%"))
		qr = qr.Or("apps.key LIKE ?", fmt.Sprintf("%s%s%s", "%", search, "%"))
	}
	if len(status) > 0 {
		qr = qr.Where("apps.status = ?", status)
	}

	qr = qr.Count(&total).Order("id ASC")

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

func (b *base) GetSearchApps(c *gin.Context) {
	var data = make(model.Apps, 0)
	var search = c.Query("search")

	/* Query database */
	qr := b.DB.Model(model.App{})
	if len(search) > 0 {
		qr = qr.Or("apps.name LIKE ?", fmt.Sprintf("%s%s%s", "%", search, "%"))
		qr = qr.Or("apps.code LIKE ?", fmt.Sprintf("%s%s%s", "%", search, "%"))
		qr = qr.Or("apps.key LIKE ?", fmt.Sprintf("%s%s%s", "%", search, "%"))
	}

	qr = qr.Order("id ASC").Limit(100).Offset(0)

	if err := qr.Find(&data).Error; err == nil {
		c.JSON(http.StatusOK, data)
	} else {
		c.JSON(http.StatusInternalServerError, data)
	}
}

func (b *base) GetApp(c *gin.Context) {
	var app model.App

	var qr = b.DB.Model(app).Where("`id` = ?", c.Param("id"))
	if err := qr.First(&app).Error; err != nil && err != gorm.ErrRecordNotFound {
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
		c.JSON(http.StatusOK, app)
	}
}

func (b *base) PostApp(c *gin.Context) {
	var request AppRequest
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

	/* Validation */
	if err := request.Validate(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"code":    http.StatusUnprocessableEntity,
			"message": "Dữ liệu không hợp lệ",
			"errors":  err,
		})
		return
	} else {
		app = model.App{
			Name:      request.Name,
			Code:      request.Code,
			Key:       request.Key,
			Status:    request.Status,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	/* Create app */
	if err := b.DB.Create(&app).Error; err != nil {
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

func (b *base) PutApp(c *gin.Context) {
	var request AppRequest
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
	var qr = b.DB.Model(app).Where("`id` = ?", c.Param("id"))
	if err := qr.First(&app).Error; err != nil && err != gorm.ErrRecordNotFound {
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
		app.Name = request.Name
		app.Code = request.Code
		app.Key = request.Key
		app.Status = request.Status
		app.UpdatedAt = time.Now()
	}

	/* Update app */
	if err := b.DB.Save(&app).Error; err != nil {
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

func (b *base) DeleteApp(c *gin.Context) {
	var app model.App

	/* Find app */
	var qr = b.DB.Model(app).Where("`id` = ?", c.Param("id"))
	if err := qr.First(&app).Error; err != nil && err != gorm.ErrRecordNotFound {
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

	/* Update app */
	if err := b.DB.Delete(&app).Error; err != nil {
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
