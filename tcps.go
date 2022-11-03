package main

import (
	"bufio"
	"cron-test/proto"
	"fmt"
	"io"
	"net"
	"time"
)

func read(conn net.Conn) {
	fmt.Printf("exec")
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		msg, err := proto.Decode(reader)
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Println("decode msg failed, err:", err)
			return
		}
		fmt.Printf("Recived from client,data:%s\n", msg)
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

func ping(conn net.Conn) {
	for {
		time.Sleep(3 * time.Second)
		msg := "ping..."
		d := W(conn, msg)
		if d != true {
			fmt.Println("ping over:", conn)
			break
		}
	}
}

func main() {
	// 监听TCP 服务端口
	listener, err := net.Listen("tcp", "0.0.0.0:10087")
	if err != nil {
		fmt.Println("Listen tcp server failed,err:", err)
		return
	}

	for {
		// 建立socket连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Listen.Accept failed,err:", err)
			continue
		} else {
			fmt.Println("new client:", conn)

		}
		// 业务处理逻辑
		go read(conn)
		// go solveInput(conn)
		// go ping(conn)
	}
}
