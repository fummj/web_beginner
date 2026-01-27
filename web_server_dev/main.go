package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

// TCPServer
func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connetion: ", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// connをflowの最後に必ずClose()させる。
	defer conn.Close()

	reader := bufio.NewReader(conn)
	status, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error: ", err)
		response := "HTTP/1.1 400 Bad Request\r\n" + "Content-type: text/plain\r\n" + "\r\n" + "For now, let's just say 400."
		conn.Write([]byte(response))
	} else {
		fmt.Println("status: ", status)
		response := "HTTP/1.1 200 OK\r\n" + "Content-type: text/plain\r\n" + "\r\n" + "recieved your msg."
		conn.Write([]byte(response))
	}
}
