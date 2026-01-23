package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
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

func NewClient() *Client {
	return &Client{}
}

// targetとportをstdInから受け付ける。
func (c *Client) recvTargetInfo() {

	// fmt.Println("waiting for your input(e.g. hostname port)...")
	// scanner := bufio.NewScanner(os.Stdin)
	// scanner.Scan()
	//
	// target, port, _ := strings.Cut(scanner.Text(), " ")

	// c.target = target
	// c.port = port
	c.target = "127.0.0.1"
	c.port = "8080"
	c.address = fmt.Sprintf("%s:%s", c.target, c.port)

	fmt.Printf("set target=%s, port=%s \n", c.target, c.port)
}

// ターゲットとポート番号のバリデーションのエントリ。
func (c Client) validateInputArgs() error {
	// hostname, ip-address, port
	if err := c.validateTargetInfo(); err != nil {
		return err
	}

	return nil
}

// ターゲットとポート番号のバリデーション。
func (c Client) validateTargetInfo() error {
	// hostname
	if r, err := validate(matchHostNameString, c.target, nonEligibleEndpoint); !r {
		// ip-address
		r, err = validate(matchIPAddressString, c.target, nonEligibleEndpoint)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// port
	if _, err := validate(matchPortString, c.port, nonEligiblePortNumber); err != nil {
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
		requestHeader      = c.target
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
	// ↓全て取得できたことを確認。
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("resp: ", resp)
	// ヘッダーやステータスを表示
	fmt.Println("Status:", resp.Status)
	fmt.Println("Headers:", resp.Header)

	// ボディ（中身）を全て読み込む
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Body:", string(body))
	return nil
}
