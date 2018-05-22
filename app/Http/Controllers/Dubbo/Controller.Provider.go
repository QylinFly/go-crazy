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
	// "net"
	"time"
	// "fmt"
	// "net/url"
	"strings"
	// "runtime"
	"strconv"
	. "github.com/xoxo/crm-x/Config"
	"github.com/xoxo/crm-x/util/logger"
	Request "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Request"
	Etcd "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Etcd"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/LoadBalancing"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Tcp/Server"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Tcp/Client"
 )

//  var (
// 	rootPath string = "dubbomesh"
// 	serviceName string = "com.alibaba.dubbo.performance.demo.provider.IHelloService"
// 	etcdKey string = fmt.Sprintf("/%s/%s",rootPath,serviceName)
// )

type ProviderAgent struct {
	superAgent      *Request.SuperHttpClient
	etcdRegistry   	*Etcd.EtcdRegistry
	loadBalancing   *LoadBalancing.LoadBalancingCtrl
	tcpClientAgent map[*TcpServer.Client]int
	tcpClientProvide *TcpClient.Connection	
 }

func InitProvider() {

	logger.Info("in InitProvider")

	etcdUrl := Config.EtcdUrl
	etcdUrl = strings.Replace(Config.EtcdUrl,"http://","",-1)
	// 初始化
	agent := &ProviderAgent{
		etcdRegistry  : Etcd.NewClient([]string{etcdUrl}),
		tcpClientAgent : make(map[*TcpServer.Client]int),
	}

	port,_ := strconv.Atoi(Config.Port)
	agent.etcdRegistry.Register(rootPath,serviceName,port)

	agent.InitTcpClient()
	agent.InitTcpServer()
}

func (self *ProviderAgent) InitTcpServer() {

	logger.Info("in InitTcpServer")
	
	server := TcpServer.New(":" + Config.Port)

	server.OnNewClient(func(c *TcpServer.Client) {
		// new client connected
		self.tcpClientAgent[c] = server.ConnCount()
	})

	// var senLeng int = 0
	server.OnNewMessage(func(c *TcpServer.Client, message *[]byte) {
		// 数据直接转发
		Config.Time.T2 = time.Now().UnixNano()

		if len(*message)>0{
			// senLeng+=len(*message)
			// logger.Info("接收"+strconv.Itoa(senLeng))
			self.tcpClientProvide.Write(*message)
		}else{
			c.SendBytes(*message)
		}
	})
	server.OnClientConnectionClosed(func(c *TcpServer.Client, err error) {
		// connection with client lost
		logger.Info("connection with client lost")
		delete(self.tcpClientAgent, c)
	})
	
	server.Listen()
}

func (self *ProviderAgent) InitTcpClient() {

	logger.Info("in InitTcpClient")

	address :="127.0.0.1:"+strconv.Itoa(Config.DubboPort)
	self.tcpClientProvide  = TcpClient.New(address)
	self.tcpClientProvide.OnOpen(func() {
		logger.Info("agent.tcpClient.OnOpen:"+address)
	})
	self.tcpClientProvide.OnError(func(err error) {
		// if !client.Connected {
		logger.Info("agent.tcpClient.OnError:"+address)
	})
	// var senLeng int = 0
	
	self.tcpClientProvide.OnMessage(func(message []byte) {
		// 遍历map
		Config.Time.T3 = time.Now().UnixNano()

		if len(message) > 0 {
			// senLeng+=len(message)
			// logger.Info("provide发送"+strconv.Itoa(senLeng))

			for client, _ := range self.tcpClientAgent {
				client.SendBytes(message)
				return
			}
		}else{			
			self.tcpClientProvide.Write(message)
		}

	})

	go self.tcpClientProvide.Connect()

	time.Sleep(time.Second)
}