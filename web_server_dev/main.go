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
		defer conn.Close()

		if err != nil {
			fmt.Println("Error accepting connetion: ", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	status, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Println("status: ", status)
	response := "HTTP/1.1 200 OK\r\n" + "Content-type: text/plain\r\n" + "\r\n" + "recieved your msg."
	conn.Write([]byte(response))
}
