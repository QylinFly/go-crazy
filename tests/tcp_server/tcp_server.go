package main

import (
	"bufio"
	"log"
	"net"
)

// Client holds info about connection
type Client struct {
	conn   net.Conn
	Server *server
}

// TCP server
type server struct {
	address                  string // Address to open connection: localhost:9999
	onNewClientCallback      func(c *Client)
	onClientConnectionClosed func(c *Client, err error)
	onNewMessage             func(c *Client, message *[]byte)
}

// Read client data from channel
func (c *Client) listen() {
	reader := bufio.NewReader(c.conn)
	for {

		buf := make([]byte, 1024)
		num, err := reader.Read(buf)

		if err != nil {
			c.conn.Close()
			c.Server.onClientConnectionClosed(c, err)
			return
		}

		if num > 0{
			mensagem := make([]byte, num)
			copy(mensagem, buf)
	
			c.Server.onNewMessage(c,&mensagem)
		}



		// // message := new([]byte)
		// message, err := reader.Peek(16)
		// // message, err := reader.ReadString('\n')
		// if err != nil {
		// 	c.conn.Close()
		// 	c.Server.onClientConnectionClosed(c, err)
		// 	return
		// }
		// if len(message) > 0 {
		// 	c.Server.onNewMessage(c, &message)
		// }
	}
}

// Send text message to client
func (c *Client) Send(message string) error {
	_, err := c.conn.Write([]byte(message))
	return err
}

// Send bytes to client
func (c *Client) SendBytes(b []byte) error {
	_, err := c.conn.Write(b)
	return err
}

func (c *Client) Conn() net.Conn {
	return c.conn
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Called right after server starts listening new client
func (s *server) OnNewClient(callback func(c *Client)) {
	s.onNewClientCallback = callback
}

// Called right after connection closed
func (s *server) OnClientConnectionClosed(callback func(c *Client, err error)) {
	s.onClientConnectionClosed = callback
}

// Called when Client receives new message
func (s *server) OnNewMessage(callback func(c *Client, message *[]byte)) {
	s.onNewMessage = callback
}

// Start network server
func (s *server) Listen() {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Panicln(err)
		log.Fatal("Error starting TCP server.")
	}
	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		client := &Client{
			conn:   conn,
			Server: s,
		}
		go client.listen()
		s.onNewClientCallback(client)
	}
}

// Creates new tcp server instance
func New(address string) *server {
	log.Println("Creating server with address", address)
	server := &server{
		address: address,
	}

	server.OnNewClient(func(c *Client) {})
	server.OnNewMessage(func(c *Client, message *[]byte) {})
	server.OnClientConnectionClosed(func(c *Client, err error) {})

	return server
}

func main() {
	server := New("127.0.0.1:1234")

	server.OnNewClient(func(c *Client) {
		// new client connected
		// lets send some message
		
		c.Send("Hello")
	})
	server.OnNewMessage(func(c *Client, message *[]byte) {
		// new message received
		
		err := c.SendBytes(*message)
		if err != nil {
			log.Println(err)
		}
	})
	server.OnClientConnectionClosed(func(c *Client, err error) {
		// connection with client lost
		log.Println("connection with client lost")
	})

	server.Listen()
}