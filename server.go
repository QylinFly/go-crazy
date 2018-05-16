package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
	"os/signal"
	"github.com/jinzhu/configor"
	Gin "github.com/gin-gonic/gin"
	"github.com/xoxo/crm-x/routes"
	. "github.com/xoxo/crm-x/Config"
	"github.com/xoxo/crm-x/util/logger"
)




func main() {
	// load config from file
	configor.Load(&Config, ".env.yml")
	// fmt.Printf("config: %#v\n\n\n", Config)
	
	// init path
	InitPath()

	// init logger
	InitLogger()

	// init database
	InitDB()

	// init gin engine
	engine := Gin.Default()
	Route.SetupRouter(engine)

	//startNormal(engine)
	startGracefulShutdown(engine)
}
func startNormal(engine *Gin.Engine)  {
	// Listen and Server in Config.Port
	engine.Run(":"+Config.Port)
}

func startGracefulShutdown(engine *Gin.Engine)  {
	// graceful-shutdown
	var pid = os.Getpid()
	ps, _ := os.FindProcess(pid)

	// shutdown this app
	// todo add Permissions、clear e.t.c
	engine.GET("/down", func(c *Gin.Context)  {
		c.String(200, "Down ok!")
		ps.Signal(os.Interrupt)
	})

	srv := &http.Server{
		Addr:    ":"+Config.Port,
		Handler: engine,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			logger.Info(fmt.Sprintf("listen: %s\n", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Info("Shutdown Server ...")

	// Recycle
	clearAll()
	logger.Info("------------------    Recycle   ------------------")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Info(fmt.Sprintf("Server Shutdown: %s\n", err))
	}
	logger.Info("------------------Server exiting------------------")
}

func clearAll()  {
	CloseDB()
}