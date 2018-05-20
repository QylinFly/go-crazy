/**
 * 基本的app配置 包含中间件 配置等
 * File: app.go 
 * Author: QylinFly (18612116114@163.com)
 * Created: 星期 2, 2017-12-19 6:27:12 pm
 * -----
 * Modified By: QylinFly (18612116114@163.com>)
 * Modified: 星期 2, 2017-12-19 7:02:17 pm
 * -----
 * Copyright 2017 - 2027 乐编程, 乐编程
 */

 package Config

 import(
	"os"
	"log"
	"strings"
	"path/filepath"
	Gin "github.com/gin-gonic/gin"
	// "github.com/xoxo/crm-x/app/Http/Middleware"
 )

 ////////////////Middleware begin//////////////////////
 func SetupGlobalMiddleware(engine *Gin.Engine)  {
	// Global middleware
	// engine.Use(middleware.FormatResponse())
	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	// engine.Use(Gin.Recovery())
 }

 func SetupApiMiddleware(router *Gin.RouterGroup)  {
	// router.Use(middleware.FormatResponse())
	// router.Use(Gin.Recovery())
 }

 func SetupAdminMiddleware(router *Gin.RouterGroup)  {
	// router.Use(middleware.FormatResponse())
	// router.Use(Gin.Recovery())
	router.Use(func(context *Gin.Context) {
		context.Next()
	})
 }
 ////////////////Middleware end//////////////////////

var Path = struct {
	App 		string `default:"./"`
	Base 		string `default:"./"`
	Static 		string `default:"./static/"`	
	Resource	string `default:"./static/"`
	Storage		string `default:"./storage"`
	LogsDir		string `default:"./storage"`
}{}

func InitPath()  {
	var base = getCurrentDirectory()

	Path.App = base 
	Path.Base = base 
	Path.Static = base + "/static/"
	Path.Resource = base + "/static/"
	Path.Storage = base + "/storage"
	Path.LogsDir = base + "/storage"
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}