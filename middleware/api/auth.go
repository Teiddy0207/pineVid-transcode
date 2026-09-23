package api

import (
	"fmt"
	"net/http"
	"os"
	"sonla/transcode/model"
	"sonla/transcode/util"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Auth(DB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth, err := c.Cookie("auth"); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Chưa đăng nhập",
			})
			return
		} else {
			if token, err := jwt.Parse(auth, func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("APP_KEY")), nil
			}); err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "Chưa đăng nhập",
				})
				c.SetCookie("auth", "", 0, "", "", false, true)
				return
			} else {
				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					var admin model.Admin
					var id, _ = strconv.ParseInt(fmt.Sprintf("%v", claims["id"]), 10, 64)
					if err := DB.Where("`id`  = ?", id).First(&admin).Error; err != nil && err != gorm.ErrRecordNotFound {
						util.Logger.Error("Auth", zap.Error(err))
						c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
							"message": "Lỗi truy vấn cơ sở dữ liệu",
						})
						return
					} else if err == gorm.ErrRecordNotFound {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
							"message": "Tài khoản không tồn tại",
						})
						c.SetCookie("auth", "", 0, "", "", false, true)
						return
					} else {
						c.Set("admin", admin)
						c.Next()
					}
				} else {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"message": "Phiên đăng nhập đã hết hạn",
					})
					c.SetCookie("auth", "", 0, "", "", false, true)
					return
				}
			}
		}
	}
}
