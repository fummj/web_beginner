package main

import (
	"fmt"
	"os"
)

func main() {
	t, p, a := recvTargetInfo()
	client := NewClient(t, p, a)

	b, err := client.connectTCPServer()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	resp := NewResponse(b)
	// resp.Status()
	// resp.Header()
	resp.Body()

}
