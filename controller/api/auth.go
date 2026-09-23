package api

import (
	"net/http"
	"os"
	"sonla/transcode/model"
	"sonla/transcode/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (b *base) PostLogin(c *gin.Context) {
	var req LoginRequest
	var admin model.Admin
	var exp = time.Now().AddDate(0, 0, 1).Unix()

	if err := c.MustBindWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Yêu cầu không hợp lệ",
		})
		return
	}

	if err := b.DB.Where("username = ?", req.Username).First(&admin).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Lỗi truy vấn cơ sở dữ liệu",
			"error":   err,
		})
		return
	} else if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Tên đăng nhập hoặc mật khẩu không đúng",
		})
		return
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Tên đăng nhập hoặc mật khẩu không đúng",
			})
			return
		} else if req.Remember {
			exp = time.Now().AddDate(0, 0, 7).Unix()
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":       admin.ID,
			"remember": req.Remember,
			"exp":      exp,
		})
		if signed, err := token.SignedString([]byte(os.Getenv("APP_KEY"))); err != nil {
			util.Logger.Error("PostLogin", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Lỗi hệ thống",
			})
			return
		} else {
			c.SetCookie("auth", signed, int(exp), "", "", false, true)
			c.JSON(http.StatusOK, gin.H{
				"message": "Đăng nhập thành công",
			})
			return
		}
	}
}

func (b *base) PostLogout(c *gin.Context) {
	c.SetCookie("auth", "", -3600, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng xuất thành công",
	})
	return
}
