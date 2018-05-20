package TcpClient

import (
	"bufio"
	"net"
	"time"
	"strconv"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Protocol"
)

type Connection struct {
	onOpenCallback    func()
	onMessageCallback func(message []byte)
	onErrorCallback   func(err error)

	Conn      net.Conn
	Address   string
	Connected bool
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

func (self *Connection) Close() {
	self.Conn.Close()
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

func (self *Connection) read() {
	reader := bufio.NewReader(self.Conn)

	for {
		buf := make([]byte, 1024)
		num, err := reader.Read(buf)

		if err != nil {
			self.Close()
			self.onErrorCallback(err)
			return
		}
		if num > 0{
			mensagem := make([]byte, num)
			copy(mensagem, buf)
			self.onMessageCallback(mensagem)
		}
	}
}

func New(address string) *Connection {
	client := &Connection{Address: address, Connected: false}

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
				data := dubbo.GetEncoderData("linfeng")
				client.Write(data.Bytes())
			}
		}()
	})

	client.OnMessage(func(message []byte) {
		rpcResponseArray := []*Dubbo.RpcResponse{}
		res,err := Dubbo.GetDecoderData(&message,rpcResponseArray)
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