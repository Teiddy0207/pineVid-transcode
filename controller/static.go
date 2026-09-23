package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path"
	"strings"
)

func Static(c *gin.Context) {
	if strings.HasSuffix(c.Request.URL.Path, "/") {
		http.NotFound(c.Writer, c.Request)
		return
	}

	var export = os.Getenv("DVR_EXPORT")
	if len(export) == 0 {
		export = "public"
	}

	c.Header("Content-Disposition", "attachment; filename=" + path.Base(c.Request.URL.Path))
	c.Header("Content-Type", c.GetHeader("Content-Type"))

	fileServer := http.FileServer(gin.Dir(export, false))
	fileServer = http.StripPrefix("/public", fileServer)
	fileServer.ServeHTTP(c.Writer, c.Request)
}
