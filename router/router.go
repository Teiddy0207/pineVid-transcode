package router

import (
	"net/http"
	"sonla/transcode/controller"
	"sonla/transcode/controller/api"
	mwAPI "sonla/transcode/middleware/api"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Init(router *gin.Engine, db *gorm.DB) {
	var apiController = api.NewController(db)
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	/* */
	router.POST("/api/job/create", apiController.PostJobCreate)
	router.POST("/api/job/retry", apiController.PostJobRetry)

	/* */
	backendAPI := router.Group("/api/backend")
	backendAPI.POST("/auth/login", apiController.PostLogin)
	backendAPI.Use(mwAPI.Auth(db))
	backendAPI.POST("/auth/logout", apiController.PostLogout)
	backendAPI.GET("/profile", apiController.GetProfile)
	backendAPI.PATCH("/profile", apiController.PatchProfile)
	backendAPI.PATCH("/profile/password", apiController.PatchPassword)

	/* App */
	backendAPI.GET("/apps", apiController.GetApps)
	backendAPI.GET("/apps/search", apiController.GetSearchApps)
	backendAPI.POST("/apps", apiController.PostApp)
	backendAPI.GET("/app/:id", apiController.GetApp)
	backendAPI.PUT("/app/:id", apiController.PutApp)
	backendAPI.DELETE("/app/:id", apiController.DeleteApp)

	/* Job */
	backendAPI.GET("/jobs", apiController.GetJobs)
	backendAPI.POST("/jobs", apiController.PostJob)
	backendAPI.GET("/job/:id", apiController.GetJob)
	backendAPI.PUT("/job/:id", apiController.PutJob)
	backendAPI.DELETE("/job/:id", apiController.DeleteJob)

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.GET("/public/*filepath", controller.Static)
}
