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
	// . "github.com/xoxo/crm-x/util"
	. "github.com/xoxo/crm-x/Config"
	Gin "github.com/gin-gonic/gin"
	"github.com/xoxo/crm-x/util/logger"
	Request "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Request"
	Etcd "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Etcd"
 )

 var (
	endpoint_index int
	max  int
)


 type BubboAgent struct {
	superAgent       *Request.SuperHttpClient
	etcdRegistry   	*Etcd.EtcdRegistry
	endpoints  []*Etcd.Endpoint
 }

 func SetupDubbo(router *Gin.Engine) *BubboAgent {

	fmt.Println("Init SetupDubbo")

	// 初始化
	agent := &BubboAgent{
		superAgent:       Request.NewClient(),
		etcdRegistry: Etcd.NewClient([]string{Config.EtcdUrl}),
		endpoints : []*Etcd.Endpoint{},
	}

	rootPath := "dubbomesh";
	serviceName := "com.alibaba.dubbo.performance.demo.provider.IHelloService"
	etcdKey := fmt.Sprintf("/%s/%s",rootPath,serviceName)
	
	agent.endpoints = agent.etcdRegistry.Find(etcdKey)

	// Update loop
	go func() {
		defer func() {
			if e := recover(); e != nil {
				//做异常处理
				logger.Info("----异常 EtcdRegistry.Find----")
			}
		}()
		for {
			time.Sleep(5000 * time.Millisecond)
			logger.Info(Config.EtcdUrl+"----定时发现----")
			agent.endpoints = agent.etcdRegistry.Find(etcdKey)
		}
	}()

	agent.InitAgent()

	endpoint_index = 0
	
	router.Any("/", agent.Ping)
	return agent
 }

 func (agent *BubboAgent) InitAgent() {
	// https://serholiu.com/go-http-client-keepalive
	agent.superAgent.Transport.DisableKeepAlives = false
	agent.superAgent.Transport.MaxIdleConnsPerHost = 40

	agent.superAgent.Transport.MaxIdleConns=100

	agent.superAgent.Transport.Dial = func(network, addr string) (net.Conn, error) {
		dial := net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 90 * time.Second,
		}
		conn, err := dial.Dial(network, addr)
		if err != nil {
			return conn, err
		}
		fmt.Println("connect done, use", conn.LocalAddr().String())
		return conn, err
	}
 }

 func (agent *BubboAgent) GetEndpoint()  string{
	endpoint_index = (endpoint_index+1)%3
	return agent.endpoints[endpoint_index].ToString()
 }
 func (agent *BubboAgent) getParam(c *Gin.Context, key string) string{
	value := c.PostForm("key")
	if value == "" { 
		value = c.Query(key) 
	} 
	return value
 }
 func (agent *BubboAgent) Ping(c *Gin.Context )  {
	// Ping test
	request := agent.superAgent;

	// inter := agent.getParam(c, "interface")
	// method:= agent.getParam(c, "method")
	// parameterTypesString:= agent.getParam(c, "parameterTypesString")
	// parameter:= agent.getParam(c, "parameter")
	
	// args := "interface="+inter+"&method"+method+"&parameterTypesString"+parameterTypesString+"&parameter"+parameter

	_, body, _ := request.Post(agent.GetEndpoint()).
	// _, body, _ := request.Post("http://10.99.2.116:30000").
	Send("interface=com.alibaba.dubbo.performance.demo.provider.IHelloService&method=hash&parameterTypesString=Ljava%2Flang%2FString%3B&parameter=wdeadsadasdadadadasdasdasdasdasdffgfjffhgfgfddfggfd").
	Type("urlencoded").
	// Send(args).
	End()

	c.String(200,body)

 }