package Switch

import (
	"github.com/xoxo/crm-x/util/logger"
)

type Switch struct{
	name	string
	numberGoroutines int
	bufferCount int
	tasks  chan *[]byte
	workerSelf bool

	onMessageCallback func(message *[]byte)
	// sendMessageCallback   func(message *[]byte)
}

// 接收消息
func (self *Switch) OnMessage(f func(message *[]byte)) {
	self.onMessageCallback = f
}
// 发送消息
// func (self *Switch)  SendMessage(f func(message *[]byte)) {
// 	self.sendMessageCallback = f
// }

// 放入一个待处理的值
func (self *Switch) DoSendMessage( message *[]byte) {
	self.tasks <- message
}

// 
func (self *Switch) GetTasks() chan *[]byte{
	return self.tasks
}

func (self *Switch) worker(tasks chan *[]byte, worker int) {
	for {
		select {
		case task,ok := <-tasks:
			if !ok{
				logger.AppendInfo("通道被关闭:",worker)
			}
			// 处理数据
			self.onMessageCallback(task)
		default:
		}
	}
}

func (self *Switch) Worker(reqID int) {
	for {
		select {
		case task,_ := <-self.tasks:
			// 处理数据
			self.onMessageCallback(task)
		default:
		}
	}
}

// workerSelf 是否自己启动worker协程 充分利用已有协程 非常重要
func New(name string,numberGoroutines int,bufferCount int,workerSelf bool) *Switch{
	// 初始化
	s := &Switch{
					name : name,
		numberGoroutines : numberGoroutines,
			 bufferCount : bufferCount,
				   tasks : make(chan *[]byte,bufferCount),
			  workerSelf : workerSelf,
	}

	// 启动worker 非workerSelfmodel
	if workerSelf == false {
		for gr:=1;gr<= numberGoroutines;gr++ {
			go s.worker(s.tasks,gr)
		}
	}

	logger.Info(name + " Switch  启动")
	return s
}