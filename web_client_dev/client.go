package main

import (
	"bufio"
	"bytes"
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

	inEligiblePortNumber = "不適切なポート番号です。"
	inEligibleTarget     = "不適切な接続先名です。"
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

type Response struct {
	rawContent []byte
	_status    string
	_header    string
	_body      string
}

func NewResponse(buffer []byte) *Response {
	return &Response{rawContent: buffer}
}

// HTTPステータスラインを取得する。すでに持っていればその値を返す。
func (resp *Response) Status() string {
	if resp._status != "" {
		return resp._status
	}

	// rawContentからstatusの内容を取得する。
	i := bytes.Index(resp.rawContent, []byte("\r\n"))
	resp._status = string(resp.rawContent[:i])

	fmt.Println(resp._status)
	return resp._status
}

// HTTPレスポンスヘッダーを取得する。すでに持っていればその値を返す。
func (resp *Response) Header() string {
	if resp._header != "" {
		return resp._header
	}

	// rawContentからheaderの内容を取得する。
	i := bytes.Index(resp.rawContent, []byte("\r\n"))     // heaaderとstatusの境目
	j := bytes.Index(resp.rawContent, []byte("\r\n\r\n")) // headerとbodyの境目
	resp._header = string(resp.rawContent[i:j])

	fmt.Println(resp._header)
	return resp._header
}

// HTTPレスポンスボディを取得する。すでに持っていればその値を返す。
func (resp *Response) Body() string {
	if resp._body != "" {
		return resp._body
	}

	// rawContentからbodyの内容を取得する。
	i := bytes.Index(resp.rawContent, []byte("\r\n\r\n")) // headerとbodyの境目
	rawBody := resp.rawContent[i:]

	crlf := []byte("\r\n")
	rawBody = bytes.Trim(rawBody, string(crlf))
	j := bytes.Index(rawBody, crlf) + len(crlf) + 1 // chunke-sizeの境目

	k := bytes.LastIndex(rawBody, crlf) // 末尾の0との境目
	resp._body = string(rawBody[j:k])

	fmt.Println(resp._body)
	return resp._body
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
	if r, err := validate(matchHostNameString, target, inEligibleTarget); !r {
		// ip-address
		r, err = validate(matchIPAddressString, target, inEligibleTarget)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// port
	if _, err := validate(matchPortString, port, inEligiblePortNumber); err != nil {
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

func (c Client) connectTCPServer() ([]byte, error) {
	var (
		requestLine        = "GET / HTTP/1.1"
		requestHeader      = fmt.Sprintf("Host: %s \r\nConnection: close", c.target)
		httpRequestMessage = []byte(fmt.Sprint(requestLine, "\r\n", requestHeader, "\r\n", "\r\n"))
	)
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return []byte{}, err
	}

	defer conn.Close()

	_, err = conn.Write(httpRequestMessage)
	if err != nil {
		return []byte{}, err
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
			return []byte{}, err
		}
		buffer = append(buffer, slice[:n]...)
	}

	// fmt.Println("response-content: ", string(buffer))
	return buffer, nil
}
