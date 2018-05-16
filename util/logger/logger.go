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

 package logger

 import (
	"fmt"
	"go.uber.org/zap"
 )


var logger *zap.Logger

func _default(format string, a ...interface{})  {
	fmt.Printf(format, a...)
}

var Fatal = logger.Fatal
var Error = logger.Error
var Debug = logger.Debug
var Info = logger.Info
var Warn = logger.Warn

func SetLogger(log *zap.Logger)  {
	logger = log
	
	Fatal = logger.Fatal
	Error = logger.Error
	Debug = logger.Debug
	Info = logger.Info
	Warn = logger.Warn
}

func AppendError(format string, a ...interface{})  {
	msg := fmt.Sprintf(format, a...)
	logger.Error(msg)
}
func AppendInfo(format string, a ...interface{})  {
	msg := fmt.Sprintf(format, a...)
	logger.Info(msg)
}
func AppendDebug(format string, a ...interface{})  {
	msg := fmt.Sprintf(format, a...)
	logger.Debug(msg)
}
func AppendWarn(format string, a ...interface{})  {
	msg := fmt.Sprintf(format, a...)
	logger.Warn(msg)
}
func AppendFatal(format string, a ...interface{})  {
	msg := fmt.Sprintf(format, a...)
	logger.Fatal(msg)
}

