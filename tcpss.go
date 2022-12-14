package main

import (
	"bufio"
	"cron-test/proto"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"io"
	"net"
	"time"
)

// ClientManager 客户端管理
type ClientManager struct {
	clients    map[*Client]bool //客户端 map 储存并管理所有的长连接client，在线的为true，不在的为false
	broadcast  chan []byte      //web端发送来的的message我们用broadcast来接收，并最后分发给所有的client
	register   chan *Client     //新创建的长连接client
	unregister chan *Client     //新注销的长连接client
}

type Client struct {
	id   int
	uuid uuid.UUID
	conn net.Conn
	send chan []byte
}

var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register: //如果有新的连接接入,就通过channel把连接传递给conn
			fmt.Println("Register uuid", conn.uuid)
			manager.clients[conn] = true
			fmt.Println("Client quantity", len(manager.clients))
		case conn := <-manager.unregister: //断开连接时
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
			}
			fmt.Println("Disconnected", conn.uuid)
			fmt.Println("Client quantity ", len(manager.clients))
			return
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:10087")
	if err != nil {
		fmt.Println("Listen tcp server failed,err:", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Listen.Accept failed,err:", err)
			continue
		} else {
			client := &Client{
				uuid: uuid.Must(uuid.NewV4()),
				conn: conn,
				send: make(chan []byte),
			}
			go manager.start()
			manager.register <- client
			go client.Read()
			go client.Write()
			go client.Ping()
		}
	}
}

type Message struct {
	Type     int
	Describe string
	Content  string
	Pid      int
	Sid      int
}

func (c *Client) Read() {
	defer c.conn.Close()
	reader := bufio.NewReader(c.conn)
	for {
		msg, err := proto.Decode(reader)
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Println("decode msg failed, err:", err)
			return
		}
		var Msg Message
		err = json.Unmarshal([]byte(msg), &Msg)
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Println("Recived from client,data", Msg)
		if Msg.Type == 2 {
			c.id = Msg.Pid
			go c.Operations(c.id)
		}
	}
}

func (c *Client) Operations(id int) {
	for conn := range manager.clients {
		fmt.Println("conn.id", conn.id)
		if conn.id == id {
			conn.send <- []byte("action!")
		}
	}
	time.Sleep(3 * time.Second)
	c.Operations(id)
}

func (c *Client) Write() {
	defer c.conn.Close()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.Write([]byte{})
				return
			}
			_, err := c.conn.Write([]byte(message))
			if err != nil {
				fmt.Println("Server Write failed,err:", err)
				return
			}
		}
	}
}

func W(conn net.Conn, msg string) bool {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Server Write failed,err:", err)
		return false
	}
	return true
}

func (c *Client) Ping() {
	for {
		time.Sleep(3 * time.Second)
		msg := "ping ==>" + c.uuid.String()
		d := W(c.conn, msg)
		if d != true {
			manager.unregister <- c
			break
		}
	}
}
