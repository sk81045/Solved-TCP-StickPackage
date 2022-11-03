package main

import (
	"cron-test/proto"
	"encoding/json"
	"log"
	"net"
	"time"
)

var (
	stop = make(chan string)
)

// ClientManager 客户端管理
type ClientManager struct {
	clients    map[*Client]bool //客户端 map 储存并管理所有的长连接client，在线的为true，不在的为false
	broadcast  chan []byte      //web端发送来的的message我们用broadcast来接收，并最后分发给所有的client
	register   chan *Client     //新创建的长连接client
	unregister chan *Client     //新注销的长连接client
}

type Client struct {
	conn net.Conn
	send chan []byte
}

//创建客户端管理者
var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func main() {
	manager.MomentConn("10087")
	for {

	}
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register: //如果有新的连接接入,就通过channel把连接传递给conn
			log.Println("Create Silce", conn)
			manager.clients[conn] = true
		case conn := <-manager.unregister: //断开连接时
			log.Println("断开连接", conn)
			stop <- "i"
			manager.MomentConn("10087")
			return
		}
	}
}

func (manager *ClientManager) MomentConn(port string) {
	if port == "" {
		port = "10087"
	}
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		log.Println("Repeat Connect,err:", err)
		time.Sleep(time.Second * 2)
		manager.MomentConn("10087")
	} else {
		client := &Client{
			conn: conn,
			send: make(chan []byte),
		}
		go manager.start()
		manager.register <- client
		log.Println("Connect Succuses...")
		go client.Read()
		go client.Ping()
		go client.Register()
	}
}

func (c *Client) Read() {
	conn := c.conn
	log.Printf("Read to Run...")
	defer conn.Close()
	for {
		var buf [128]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			conn.Close()
			manager.unregister <- c
			log.Println("检测到服务端断开,err:", err)
			return
		}
		data := string(buf[:n])
		// if data == "ping..." {
		// 	// W(conn, "Received ping...")
		// }
		log.Printf("Recived from serve,data:%s\n", data)
	}
}

type Message struct {
	Type     int
	Describe string
	Content  string
	Pid      int
	Sid      int
}

func (c *Client) Register() {
	defer c.conn.Close()
	item := &Message{
		Type:     2,
		Describe: "register",
		Content:  "Succuses",
		Pid:      1045,
		Sid:      888,
	}
	msg, err := json.Marshal(item)
	if err != nil {
		panic(err)
	}
	for {
		time.Sleep(2 * time.Second)
		d := W(c.conn, string(msg))
		if d != true {
			log.Println("ping over:", c.conn)
			break
		}
		select {}
	}
}

func (c *Client) Ping() {
	defer c.conn.Close()
	item := &Message{
		Type:     1,
		Describe: "ping",
		Content:  "Succuses",
		Pid:      2698,
		Sid:      888,
	}
	msg, err := json.Marshal(item)
	if err != nil {
		panic(err)
	}
	for {
		time.Sleep(2 * time.Second)
		d := W(c.conn, string(msg))
		if d != true {
			log.Println("ping over:", c.conn)
			break
		}
	}
	select {
	case <-stop:
		log.Println("ping stop:")
		return
	}
}

func W(conn net.Conn, msg string) bool {
	data, err := proto.Encode(msg)
	if err != nil {
		log.Println("encode msg failed, err:", err)
		return false
	}
	_, err = conn.Write(data)
	if err != nil {
		log.Println("写入错误:", err)
		return false
	}
	return true
}

func (c *Client) Ping() {
	for {
		time.Sleep(1 * time.Second)
		msg := "ping..."
		d := W(c.conn, msg)
		if d != true {
			log.Println("ping over:", c.conn)
			break
		}
	}
}
