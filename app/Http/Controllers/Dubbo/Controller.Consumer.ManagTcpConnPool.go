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
	"sync"
	"time"
	"strconv"
	"github.com/xoxo/crm-x/util/logger"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Util"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Tcp/Client"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Protocol"
 )


var (
	lastEndpoints *map[string]int = nil
	updateMutex  sync.Mutex
)

func (self *BubboAgent) UpdateTcpConnPool(nodes []*Util.Endpoint){
	endpoints := make(map[string]int)
	for _, node := range nodes {
		endpoints[node.ToString()] = node.Weight()
		self.tcpConnPool.Get(node.ToString())
	}
	updateMutex.Lock()
	if lastEndpoints !=nil {
		// 删除掉线用户
		for key, _ := range *lastEndpoints {
			if _, ok := endpoints[key]; !ok {  
				// 不存在  就删除
				self.tcpConnPool.RemoveNode(key)
			}  
		}
	}
	lastEndpoints = &endpoints
	updateMutex.Unlock()
}

func (self *BubboAgent) InitTcpConnPool(){
	self.tcpConnPool, _ = TcpClient.NewConnPool(func(address string) (TcpClient.ConnRes, error) {
		tcpClient  := TcpClient.New(address)//10.99.2.116:20880 or ":18080"
		tcpClient.OnOpen(func() {
			logger.Info("agent.tcpClient.OnOpen : "+address)
		})
		tcpClient.OnError(func(err error) {
			// if !client.Connected {
			logger.Info("agent.tcpClient.OnError : "+address)
		})
		// var senLeng int = 0
		tcpClient.OnMessage(func(message []byte) {
			if len(message) == 0 {
				tcpClient.Write(message)
				return
			}
			// senLeng+=len(message)
			// logger.Info("Agent接收"+strconv.Itoa(senLeng))
			rpcResponseArray := []*Dubbo.RpcResponse{}
			res,err,data:= Dubbo.GetDecoderData(&message,rpcResponseArray)
			for _, ev := range res {
				stopCh,ok := self.mapReq.Load(ev.ID())
				if !ok{
					time.Sleep(time.Millisecond)
					stopCh,ok = self.mapReq.Load(ev.ID())
				}
				if ok {
					stopCh2 := stopCh.(chan string)
					if stopCh2 !=nil{
						stopCh2 <- string(*ev.Data())
					}
				}else{
					logger.Error("Menssage agent.mapReq.Load error : " +strconv.FormatUint(ev.ID(),10)+"---"+ string(*ev.Data()))
				}
				// println("Menssage : " +strconv.FormatUint(ev.ID(),10)+"---"+ string(*ev.Data()))
			}
			if data != nil{
				tcpClient.SetLast(data)
			}
			if err != ""{
				logger.Debug("Menssage error: " + err)
			}
		})
		//  链接运行
		go tcpClient.Connect()
		time.Sleep(time.Millisecond)

        return tcpClient,nil
 	})
}