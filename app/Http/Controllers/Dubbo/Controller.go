/**
 * File: CommonController.go 通用控制方法
 * Author: QylinFly (18612116114@163.com)
 * Created: 星期 2, 2017-12-19 6:18:50 pm
 * -----
 * Modified By: QylinFly (18612116114@163.com>)
 * Modified: 星期 2, 2017-12-19 6:19:24 pm
 * -----
 * Copyright 2017 - 2027 乐编程, 乐编程
 */


 package DubboController

 import(
	"net"
	"time"
	"fmt"
	. "github.com/xoxo/crm-x/util"
	Gin "github.com/gin-gonic/gin"
	Request "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Request"
	Etcd "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Etcd"
 )

 var (
	reqs int
	max  int
)

 type BubboAgent struct {
	superAgent        *Request.SuperHttpClient
 }

 func SetupDubbo(router *Gin.Engine) *BubboAgent {

	fmt.Println("Init SetupDubbo")
	agent := &BubboAgent{
		superAgent:       Request.NewClient(),
	}

	agent.InitAgent()
	router.GET("/", agent.Ping)
	return agent
 }

 func (agent *BubboAgent) InitAgent() {
	// https://serholiu.com/go-http-client-keepalive
	agent.superAgent.Transport.DisableKeepAlives = false
	agent.superAgent.Transport.MaxIdleConnsPerHost = 40

	agent.superAgent.Transport.MaxIdleConns=100

	agent.superAgent.Transport.Dial = func(network, addr string) (net.Conn, error) {
		dial := net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		
		conn, err := dial.Dial(network, addr)
		if err != nil {
			return conn, err
		}
	
		fmt.Println("connect done, use", conn.LocalAddr().String())
	
		return conn, err
	}

 }

 func (agent *BubboAgent) Ping(c *Gin.Context )  {
	// Ping test
	request := agent.superAgent;
	_, body, _ := request.Post("http://10.99.2.116:30000").
	Type("urlencoded").
	Send("interface=com.alibaba.dubbo.performance.demo.provider.IHelloService&method=hash&parameterTypesString=Ljava%2Flang%2FString%3B&parameter=wdeadsadasdadadadasdasdasdasdasdffgfjffhgfgfddfggfd").
	End()

	c.String(200, "pong")
	Api_response(c,Gin.H{"user": body})
 }