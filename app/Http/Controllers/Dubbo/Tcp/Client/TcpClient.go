package TcpClient

import (
	"bufio"
	"net"
	"sync"
	"time"
	"strconv"
	"bytes"
	"github.com/xoxo/crm-x/util/logger"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Protocol"
)

type Connection struct {
	onOpenCallback    func()
	onMessageCallback func(message []byte)
	onErrorCallback   func(err error)

	Conn      net.Conn
	Address   string
	Connected bool
	lastMutex sync.Mutex
	last *[]byte //解析剩余 粘包导致
}

func (self *Connection) SetLast(data *[]byte) {
	self.lastMutex.Lock()
	 self.last = data
	self.lastMutex.Unlock()
}

func (self *Connection) OnOpen(f func()) {
	self.onOpenCallback = f
}

func (self *Connection) OnMessage(f func(message []byte)) {
	self.onMessageCallback = f
}

func (self *Connection) OnError(f func(err error)) {
	self.onErrorCallback = f
}

func (self *Connection) Close() error{
	return self.Conn.Close()
}

func (self *Connection) Write(message []byte) {
	self.Conn.Write(message)
}

func (self *Connection) WriteString(message string) {
	self.Conn.Write([]byte(message))
}

func (self *Connection) Connect() {
	client, err := net.Dial("tcp", self.Address)

	if err != nil {
		self.onErrorCallback(err)
	} else {
		defer client.Close()
		self.Conn = client

		self.Connected = true
		self.onOpenCallback()
		self.read()
	}
}
// 一个[]byte的对象池，每个对象为一个[]byte
var bytePool = sync.Pool{
	New: func() interface{} {
	  b := make([]byte, 1024)
	  return &b
	},
}

var buffer *bytes.Buffer = new(bytes.Buffer)

func (self *Connection) read() {
	reader := bufio.NewReader(self.Conn)

	for {
		// sync.Pool 优化 https://www.jianshu.com/p/2bd41a8f2254
		buf := bytePool.Get().(*[]byte)
		num, err := reader.Read(*buf)

		if err != nil {
			self.Close()
			self.onErrorCallback(err)
			bytePool.Put(buf)
			return
		}
		if num > 0{
			if self.last != nil{
				logger.Debug("触发数据拼接了！")
				mensagem := make([]byte, num)
				copy(mensagem, *buf)
				bytePool.Put(buf)
				
				buffer.Reset()
				self.lastMutex.Lock()
				buffer.Write(*self.last)
				self.last = nil
				self.lastMutex.Unlock()

				buffer.Write(mensagem)

				data := buffer.Bytes()
				self.onMessageCallback(data)
			}else{
				mensagem := make([]byte, num)
				copy(mensagem, *buf)
				bytePool.Put(buf)
				self.onMessageCallback(mensagem)
			}
		}
	}
}

func New(address string) *Connection {
	client := &Connection{Address: address, Connected: false,last:nil}

	client.OnOpen(func() {})
	client.OnError(func(err error) {})
	client.OnMessage(func(message []byte) {})

	return client
}


func mainTest() {
	// client := New("10.99.2.116:20880")
	client := New("127.0.0.1:1234")
	
	dubbo := Dubbo.New()
	client.OnOpen(func() {

		var idx int = 0
		go func() {
			for idx <10{
				idx++
				// println("---" + strconv.Itoa(idx))
				time.Sleep(time.Millisecond )
				_,data := dubbo.GetEncoderData("linfeng")
				client.Write(data.Bytes())
			}
		}()
	})

	client.OnMessage(func(message []byte) {
		rpcResponseArray := []*Dubbo.RpcResponse{}
		res,err,_ := Dubbo.GetDecoderData(&message,rpcResponseArray)
		if err == ""{
			for _, ev := range res {
				println("Menssage : " +strconv.FormatUint(ev.ID(),10)+"---"+ string(*ev.Data()))
			}
		}else{
			println("Menssage error: " + err)
		}
	})

	client.OnError(func(err error) {
		if !client.Connected {
			panic(err)
		} else {
			println(err.Error())
		}
	})

	client.Connect()
}