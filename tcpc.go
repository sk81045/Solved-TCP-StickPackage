package main

import (
	"bufio"
	"cron-test/tcp/proto"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var (
	ConnectStatus = false
)

func read(conn net.Conn) {
	fmt.Printf("exec")
	defer conn.Close()
	for {
		var buf [128]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			ConnectStatus = false
			// connect()
			fmt.Println("Read from tcp server failed,err:", err)
			break
		}
		data := string(buf[:n])
		if data == "ping..." {
			W(conn, "Received ping...")
		}
		fmt.Printf("Recived from serve,data:%s\n", data)
	}
}

func W(conn net.Conn, msg string) bool {
	data, err := proto.Encode(msg)
	if err != nil {
		fmt.Println("encode msg failed, err:", err)
		return false
	}

	for i := 0; i < 10; i++ {
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("Server Write failed,err:", err)
			return false
		}
	}

	return true
}

func write(conn net.Conn) {
	fmt.Printf("write")
	defer conn.Close()
	inputReader := bufio.NewReader(os.Stdin)
	// 一直读取直到遇到换行符
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Println("Read from console failed,err:", err)
			return
		}

		// 读取到字符"Q"退出
		str := strings.TrimSpace(input)
		if str == "Q" {
			break
		}

		W(conn, input)
	}
}

func connect() {
	for {
		time.Sleep(time.Second * 3)
		if ConnectStatus == false {
			fmt.Println("Restart")
			// 连接服务器
			conn, err := net.Dial("tcp", "127.0.0.1:10087")
			if err != nil {
				fmt.Println("Connect to TCP server failed ,err:", err)
				continue
			} else {
				fmt.Println("Connect Succuses")
				ConnectStatus = true
				go read(conn)
				go write(conn)
			}

		}
	}
}

func main() {
	connect()
	for {

	}
}
