package main

import (
	"bufio"
	"cron-test/proto"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var (
	R      = false //连接状态
	wirteC = make(chan string, 2)
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
			// if _, ok := manager.clients[conn]; ok {
			// 	close(conn.send)
			// 	delete(manager.clients, conn)
			// }
			wirteC <- "exit" // 入 chan
			manager.MomentConn("10087")
			return
			// case message := <-manager.broadcast: //收到服务端消息 开始广播
		}
	}
}

func (manager *ClientManager) MomentConn(port string) {
	log.Println("Beginning Connect..")
	if port == "" {
		port = "10087"
	}
	// for {
	// time.Sleep(time.Second * 2)

	log.Println("port", port)

	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		log.Println("连接失败 再次尝试,err:", err)
		time.Sleep(time.Second * 2)
		manager.MomentConn("10087")
		// continue
	} else {
		R = true
		client := &Client{
			conn: conn,
			send: make(chan []byte),
		}
		go manager.start()
		manager.register <- client
		log.Println("Connect Succuses...")
		go client.Read()
		go client.Ping()

		// continue
		// }
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
			// time.Sleep(time.Second * 1)
			log.Println("检测到服务端断开,err:", err)
			return
		}
		data := string(buf[:n])
		if data == "ping..." {
			// W(conn, "Received ping...")
		}
		log.Printf("Recived from serve,data:%s\n", data)
	}
}

func (c *Client) Write() {
	defer c.conn.Close()
	inputReader := bufio.NewReader(os.Stdin)
	// 一直读取直到遇到换行符
	for {
		log.Println("Write....", R)
		input, err := inputReader.ReadString('\n')
		log.Println("input", input)

		if err != nil {
			log.Println("Read from console failed,err:", err)
			return
		}

		// 读取到字符"Q"退出
		str := strings.TrimSpace(input)
		if str == "Q" {
			break
		}
		W(c.conn, input)
		// select {
		// //如果有新的连接接入,就通过channel把连接传递给conn
		// case <-wirteC:
		// 	log.Println("Write 停止!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		// 	break
		// }
		log.Println("blocking！！")
	}
}

func W(conn net.Conn, msg string) bool {
	data, err := proto.Encode(msg)
	if err != nil {
		log.Println("encode msg failed, err:", err)
		return false
	}
	// for i := 0; i < 10; i++ {
	_, err = conn.Write(data)
	if err != nil {
		log.Println("写入错误:", err)
		// stop <- 3
		return false
	}

	// }

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
		// select {
		// //如果有新的连接接入,就通过channel把连接传递给conn
		// case <-wirteC:
		// 	log.Println("Write 停止!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		// 	break
		// }
	}
}
