package api

import (
	"net/http"
	"sonla/transcode/model"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golang.org/x/crypto/bcrypt"
)

func (b *base) GetProfile(c *gin.Context) {
	var admin = c.MustGet("admin").(model.Admin)
	c.JSON(http.StatusOK, gin.H{
		"username":  admin.Username,
		"fullName":  admin.FullName,
		"createdAt": admin.CreatedAt.Unix(),
		"updatedAt": admin.UpdatedAt.Unix(),
	})
}

func (b *base) PatchProfile(c *gin.Context) {
	var admin = c.MustGet("admin").(model.Admin)
	var req ProfileRequest

	if err := c.MustBindWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Yêu cầu không hợp lệ",
		})
		return
	}

	if err := b.DB.Model(admin).Where("id = ?", admin.ID).Updates(map[string]interface{}{
		"full_name": req.FullName,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Lỗi truy vấn cơ sở dữ liệu",
			"error":   err,
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Cập nhật tài khoản thành công",
		})
		return
	}
}

func (b *base) PatchPassword(c *gin.Context) {
	var admin = c.MustGet("admin").(model.Admin)
	var req PasswordRequest

	if err := c.MustBindWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Yêu cầu không hợp lệ",
		})
		return
	}

	/* Validation */
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"code":    http.StatusUnprocessableEntity,
			"message": "Dữ liệu không hợp lệ",
			"errors":  err,
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Mật khẩu cũ không đúng",
		})
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Mật khẩu nhập lại không khớp",
		})
		return
	}

	if ph, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Lỗi hệ thống",
		})
		return
	} else {
		if err := b.DB.Model(admin).Where("id = ?", admin.ID).Updates(map[string]interface{}{
			"password_hash": string(ph),
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Lỗi truy vấn cơ sở dữ liệu",
				"error":   err,
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Cập nhật mật khẩu thành công",
			})
			return
		}
	}
}
