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
	// . "github.com/xoxo/crm-x/util"
	. "github.com/xoxo/crm-x/Config"
	Gin "github.com/gin-gonic/gin"
	"github.com/xoxo/crm-x/util/logger"
	Request "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Request"
	Etcd "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Etcd"
	// "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Util"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/LoadBalancing"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Protocol"
	
 )

 var (
	rootPath string = "dubbomesh";
	serviceName string = "com.alibaba.dubbo.performance.demo.provider.IHelloService"

	etcdKey string = fmt.Sprintf("/%s/%s",rootPath,serviceName)
	
)

 type BubboAgent struct {
	superAgent      *Request.SuperHttpClient
	etcdRegistry   	*Etcd.EtcdRegistry
	loadBalancing   *LoadBalancing.LoadBalancingCtrl
 }

 func SetupDubbo(router *Gin.Engine) *BubboAgent {

	logger.Info("Init SetupDubbo")

	etcdUrl := Config.EtcdUrl
	etcdUrl = strings.Replace(Config.EtcdUrl,"http://","",-1)
	// 初始化
	agent := &BubboAgent{
		superAgent    : Request.NewClient(),
		etcdRegistry  : Etcd.NewClient([]string{etcdUrl}),
		loadBalancing : nil,
	}

	// 节点发现和负载均衡初始化
	endpoints := agent.etcdRegistry.Find(etcdKey)
	agent.loadBalancing = LoadBalancing.New(endpoints)

	go agent.EndPointWatch()
	go agent.loadBalancing.RecordResponseInfoLoop()

	agent.InitAgent()
	
	router.GET("/", agent.PingV1)
	router.POST("/", agent.PingV1)

	router.GET("/info", agent.PrintInfo)
	



	Dubbo.New().InitHeader()

	return agent
 }

//  服务发现逻辑
 func (agent *BubboAgent) EndPointWatch(){
	events, err := agent.etcdRegistry.WatchTree("/"+rootPath)
	if err == nil{
		for { // Check for updates
			select {
			case event := <-events:
				logger.Info("EtcdRegistry WatchTree Events >>"+event[0].Key)
				endpoints := agent.etcdRegistry.Find(etcdKey)
				// endpoints := agent.etcdRegistry.KVPairToEndpoint(event)
				agent.loadBalancing.Update(endpoints)
			}
		}
 	}else{
		logger.Info("EndPointWatch 失败！")
		time.Sleep(5 * time.Second)
		agent.EndPointWatch()
	 }
}

 func (agent *BubboAgent) InitAgent() {
	// https://serholiu.com/go-http-client-keepalive
	agent.superAgent.Transport.DisableKeepAlives = false
	agent.superAgent.Transport.MaxIdleConnsPerHost = 260

	agent.superAgent.Transport.MaxIdleConns=100

	agent.superAgent.Transport.Dial = func(network, addr string) (net.Conn, error) {
		dial := net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 3000 * time.Second,
		}
		conn, err := dial.Dial(network, addr)
		if err != nil {
			return conn, err
		}
		logger.Debug("connect done, use "+ conn.LocalAddr().String())
		return conn, err
	}

	agent.superAgent.InitPool()

 }

 func (agent *BubboAgent) GetEndpoint()  string{
	return agent.loadBalancing.Get()
 }

 func (agent *BubboAgent) getParam(c *Gin.Context, key string) string{
	value := c.PostForm(key)
	if value == "" { 
		value = c.Query(key) 
	} 
	value = url.QueryEscape(value)
	return value
 }

 func (agent *BubboAgent) PrintInfo(c *Gin.Context )  {
	agent.loadBalancing.PrintRecordResponseInfo()
 }

 func (agent *BubboAgent) Ping(c *Gin.Context )  {

	// Ping test
	request := agent.superAgent;

	inter := agent.getParam(c, "interface")
	method:= agent.getParam(c, "method")
	parameterTypesString:= agent.getParam(c, "parameterTypesString")
	parameter:= agent.getParam(c, "parameter")
	
	args := "interface="+inter+"&method="+method+"&parameterTypesString="+parameterTypesString+"&parameter="+parameter

	// logger.Info(args)
	
	targetUrl := agent.GetEndpoint()

	stopCh, err := request.Post(targetUrl).
	Type("urlencoded").
	Send(args).
	EndPool(c)

	if err == nil{
		startTime := time.Now()
		for { // Check for updates
			select {
			case event := <-stopCh:
				// nanosecond 请求耗时
				latency := time.Since(startTime)
				latencyInt :=  int(latency.Seconds()*10000.0)
				agent.loadBalancing.RecordResponseInfo(targetUrl,latencyInt)

				logger.Info("内部耗时="+ strconv.Itoa(latencyInt-event)+"   调用耗时="+strconv.Itoa(event) )
				
				return
				
			case <-time.After(2 * time.Second):
				return
			}
		}
 	}else{
		 logger.Info("EndPool return error!")
	 }

 }

 func (agent *BubboAgent) PingV1(c *Gin.Context )  {

	request := agent.superAgent;

	inter := agent.getParam(c, "interface")
	method:= agent.getParam(c, "method")
	parameterTypesString:= agent.getParam(c, "parameterTypesString")
	parameter:= agent.getParam(c, "parameter")
	
	args := "interface="+inter+"&method="+method+"&parameterTypesString="+parameterTypesString+"&parameter="+parameter


	startTime := time.Now()

	targetUrl := agent.GetEndpoint()


	_, body, _ := request.Post(targetUrl).
	// _, body, _ := request.Post("http://10.99.2.116:30000").
	// Send("interface=com.alibaba.dubbo.performance.demo.provider.IHelloService&method=hash&parameterTypesString=Ljava%2Flang%2FString%3B&parameter=wdeadsadasdadadadasdasdasdasdasdffgfjffhgfgfddfggfd").
	Type("urlencoded").
	Send(args).
	End()

	c.String(200,body)

	// nanosecond 请求耗时
	latency := time.Since(startTime)

	latencyInt :=  int(latency.Nanoseconds()/100000.0)

	agent.loadBalancing.RecordResponseInfo(targetUrl,latencyInt)

	logger.Info("------------"+ strconv.Itoa(runtime.NumGoroutine())+"    "+strconv.Itoa(latencyInt) )

 }