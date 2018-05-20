/*
 * 通用控制方法
 * File: Controller.go
 * Author: QylinFly (18612116114@163.com)
 * Created: 星期 3, 2018-5-16 1:25:04 pm
 * -----
 * Modified By: QylinFly (18612116114@163.com>)
 * Modified: 星期 4, 2018-5-17 1:47:13 pm
 * -----
 * Copyright 2017 - 2027
 */



 package DubboController

 import(
	"net"
	"time"
	"fmt"
	"net/url"
	"strings"
	"runtime"
	"strconv"
	. "github.com/xoxo/crm-x/Config"
	Gin "github.com/gin-gonic/gin"
	"github.com/xoxo/crm-x/util/logger"
	Request "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Request"
	Etcd "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Etcd"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/LoadBalancing"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Protocol"

 )

 var (
	rootPath string = "dubbomesh"
	serviceName string = "com.alibaba.dubbo.performance.demo.provider.IHelloService"
	etcdKey string = fmt.Sprintf("/%s/%s",rootPath,serviceName)
)

type ProviderAgent struct {
	superAgent      *Request.SuperHttpClient
	etcdRegistry   	*Etcd.EtcdRegistry
	loadBalancing   *LoadBalancing.LoadBalancingCtrl
 }

func InitProvider() *ProviderAgent{
	if Config.Type != "provider"{
		return
	}

	etcdUrl := Config.EtcdUrl
	etcdUrl = strings.Replace(Config.EtcdUrl,"http://","",-1)
	// 初始化
	agent := &ProviderAgent{
		etcdRegistry  : Etcd.NewClient([]string{etcdUrl}),
	}

	etcdRegistry.Register(rootPath,serviceName,Config.Port)

	return etcdRegistry
}
