package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	httpPortNum = 8080

	matchHostNameString  = "\\.[a-z]+$"
	matchIPAddressString = "[0-9]+.[0-9]+.[0-9]+"
	matchPortString      = "^8080$"

	nonEligiblePortNumber = "不適切なポート番号です。"
	nonEligibleEndpoint   = "不適切な接続先名です。"
)

type Client struct {
	target     string
	port       string
	httpMethod string
	address    string
}

func NewClient(target, port, address string) *Client {
	return &Client{
		target:  target,
		port:    port,
		address: address,
	}
}

// targetとportをstdInから受け付ける。
func recvTargetInfo() (string, string, string) {

	fmt.Println("waiting for your input(e.g. hostname port)...")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	target, port, b := strings.Cut(scanner.Text(), " ")
	if !b {
		// test my-tcp-server
		// target = "127.0.0.1"
		// port = "8080"

		// test google
		target = "www.google.com"
		port = "80"
	}

	address := fmt.Sprintf("%s:%s", target, port)

	fmt.Printf("address: target=%s, port=%s \n", target, port)

	return target, port, address
}

// ターゲットとポート番号のバリデーションのエントリ。
func validateInputArgs(target, port string) error {
	// hostname, ip-address, port
	if err := validateTargetInfo(target, port); err != nil {
		return err
	}

	return nil
}

// ターゲットとポート番号のバリデーション。
func validateTargetInfo(target, port string) error {
	// hostname
	if r, err := validate(matchHostNameString, target, nonEligibleEndpoint); !r {
		// ip-address
		r, err = validate(matchIPAddressString, target, nonEligibleEndpoint)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// port
	if _, err := validate(matchPortString, port, nonEligiblePortNumber); err != nil {
		return err
	}

	return nil

}
func validate(m string, c string, e string) (bool, error) {
	r, err := regexp.MatchString(m, c)
	if !r {
		return r, errors.New(e)
	} else if err != nil {
		return r, err
	}

	return r, nil
}

func (c Client) connectTCPServer() error {
	var (
		requestLine        = "GET / HTTP/1.1"
		requestHeader      = fmt.Sprintf("Host: %s \r\nConnection: close", c.target)
		httpRequestMessage = []byte(fmt.Sprint(requestLine, "\r\n", requestHeader, "\r\n", "\r\n"))
	)
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.Write(httpRequestMessage)
	if err != nil {
		return err
	}

	conn.SetReadDeadline(time.Now().Add(time.Second * 5))

	slice := make([]byte, 1024)
	var buffer []byte
	var n int = 1

	for n != 0 {
		n, err = conn.Read(slice)
		if err != nil {
			if err == io.EOF {
				fmt.Println("done!")
				break
			}
			fmt.Println("can not read response error: ", err)
			return err
		}
		buffer = append(buffer, slice[:n]...)
	}

	fmt.Println("response-content: ", string(buffer))
	return nil
}
