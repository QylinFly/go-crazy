/**
 * 日志纪录配置
 * File: logger.go
 * Author: QylinFly (18612116114@163.com)
 * Created: 星期 3, 2017-12-20 11:58:42 am
 * -----
 * Modified By: QylinFly (18612116114@163.com>)
 * Modified: 星期 3, 2017-12-20 11:58:47 am
 * -----
 * Copyright 2017 - 2027 乐编程, 乐编程
 */

 package Config

 import (
	"fmt"
	"log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"github.com/xoxo/crm-x/util/logger"
 )

 var _logger *zap.Logger

 func InitLogger()  {
	var err error
	cfg := zap.NewProductionConfig()

	// config 
	cfg.OutputPaths = []string{
		"stdout",
		Path.Storage+"logs/go-crazy.log",
	}
	cfg.ErrorOutputPaths = []string{
		"stderr",
		Path.Storage+"logs/go-crazy-error.log",
	}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// level 
	cfg.Level.SetLevel(zap.DebugLevel)

	// 建立
	_logger,err = cfg.Build()
	if(err != nil){
		log.Println(fmt.Sprintf("\n Init logger error, and got err=%+v\n", err))
	}
	
	// start
	_logger.Info("--------------------------------------------------")
	_logger.Info("-------------------App start----------------------")
	_logger.Info("--------------------------------------------------")

	logger.SetLogger(_logger)
 }