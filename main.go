package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"sonla/transcode/database"
	"sonla/transcode/database/seed"
	"sonla/transcode/job"
	"sonla/transcode/middleware/cors"
	"sonla/transcode/router"
	"sonla/transcode/util"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
)

//go:embed layout/*
var f embed.FS

func main() {
	seedData := flag.Bool("s", false, "Seed data")
	flag.Parse()

	/* Init Logger */
	util.InitLogger()

	/* Connect mysql */
	db, err := database.ConnectMysql()
	if err != nil {
		util.Logger.Fatal("Mysql: connect error", zap.String("error", err.Error()))
		os.Exit(1)
	}

	/* Check command */
	if *seedData == true {
		seed.Run(db)
		return
	}

	/* Config mode */
	if os.Getenv("ENV") != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.ForwardedByClientIP = true
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	if os.Getenv("SEVER_LOG") == "true" {
		r.Use(gin.Logger())
	}

	/* */
	r.Use(cors.Middleware())

	/* Template */
	templ := template.Must(template.New("").ParseFS(f, "layout/index.html"))
	r.SetHTMLTemplate(templ)
	fAssets := fs.FS(f)
	assets, _ := fs.Sub(fAssets, "layout/assets")
	r.StaticFS("/assets", http.FS(assets))

	/* Recovery if error */
	r.Use(gin.Recovery())
	router.Init(r, db)

	go job.Run(db)

	/* Run server */
	if err := endless.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")), r); err != nil {
		util.Logger.Info("Server stop")
	}
}
