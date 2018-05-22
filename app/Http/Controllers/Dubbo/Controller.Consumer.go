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
	"sync"
	"net/url"
	"strings"
	"bytes"
	"strconv"
	. "github.com/xoxo/crm-x/Config"
	Gin "github.com/gin-gonic/gin"
	"github.com/xoxo/crm-x/util/logger"
	Request "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Request"
	Etcd "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Etcd"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/LoadBalancing"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Protocol"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Tcp/Client"
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
	rpcEncoder		*Dubbo.DubboRpcEncoder
	tcpConnPool  	*TcpClient.ConnPool
	    mapReq      sync.Map // reqID -- chan string
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
		rpcEncoder    : Dubbo.New(),
	}

	agent.InitTcpConnPool()
	
	// 节点发现和负载均衡初始化
	endpoints := agent.etcdRegistry.Find(etcdKey)
	agent.UpdateTcpConnPool(endpoints)
	agent.loadBalancing = LoadBalancing.New(endpoints)

	go agent.EndPointWatch()
	go agent.loadBalancing.RecordResponseInfoLoop()

	// agent.InitAgent()
	router.GET("/", agent.Ping)
	router.POST("/", agent.Ping)
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
					agent.UpdateTcpConnPool(endpoints)
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

//  获取编码后数据包
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

 var maxClient = make(chan byte,270)


 func (agent *BubboAgent) GetEncoderDataByContext(c *Gin.Context ) (uint64,*bytes.Buffer) {
	// Ping test
	// request := agent.superAgent;

	// inter := agent.getParam(c, "interface")
	// method:= agent.getParam(c, "method")
	// parameterTypesString:= agent.getParam(c, "parameterTypesString")
	parameter:= agent.getParam(c, "parameter")
	
	// args := "interface="+inter+"&method="+method+"&parameterTypesString="+parameterTypesString+"&parameter="+parameter

	// logger.Info(args)
	
	// targetUrl := agent.GetEndpoint()

	return agent.rpcEncoder.GetEncoderData(parameter)
 }

 func (agent *BubboAgent) Ping(c *Gin.Context )  {

	startTime := time.Now()
	
	Config.Time.T1 = time.Now().UnixNano()
	
	reqId,data := agent.GetEncoderDataByContext(c)

	maxClient <- 1
	
	targetUrl := agent.GetEndpoint()

	if targetUrl=="" {
		logger.Warn("无可用服务节点！")
		return
	}
	
	for index := 0; index < 10; index++ {
		tcpClientT,_ := agent.tcpConnPool.Get(targetUrl) // ":18080"
	
		tcpClient := tcpClientT.(*TcpClient.Connection)
	
		if tcpClient.Connected{
			tcpClient.Write(data.Bytes())
			break
		}else{
			logger.Warn("agent.tcpClient.Connected == false")
			time.Sleep(time.Millisecond*10)
		}
	}


	stopCh :=  make(chan string)

	agent.mapReq.Store(reqId,stopCh)

	time.Sleep(time.Millisecond*40)
	for { // Check for updates
		select {
		case hash := <-stopCh:
			c.String(200,hash)
			agent.mapReq.Delete(reqId)
			Config.Time.T4 = time.Now().UnixNano()
			<- maxClient
			latency := time.Since(startTime)
			latencyInt :=  int(latency.Seconds()*10000.0)
			agent.loadBalancing.RecordResponseInfo(targetUrl,latencyInt)

			logger.AppendDebug("########## agent = ",(Config.Time.T2-Config.Time.T1)/10000.0,(Config.Time.T4-Config.Time.T3)/10000.0," provide = ",(Config.Time.T3-Config.Time.T2)/10000.0) 
			return
		case <-time.After(5000 * time.Millisecond):
			logger.Info("超时返回了！"+strconv.FormatUint(reqId,10))
			agent.mapReq.Delete(reqId)
			<- maxClient
			return
		}
	}
 }
 func (agent *BubboAgent) PingV2(c *Gin.Context )  {

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

				logger.Debug("内部耗时="+ strconv.Itoa(latencyInt-event)+"   调用耗时="+strconv.Itoa(event) )
				
				return
				
			case <-time.After(2 * time.Second):
				return
			}
		}
 	}else{
		 logger.Info("EndPool return error!")
	 }

 }