package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

const (
	httpPortNum = 8080
	crlf        = "\r\n"

	matchHostNameString  = "\\.[a-z]+$"
	matchIPAddressString = "[0-9]+.[0-9]+.[0-9]+"
	matchPortString      = "^8080$"

	inEligiblePortNumber = "不適切なポート番号です。"
	inEligibleTarget     = "不適切な接続先名です。"
)

type HTTPClient struct {
	target     string
	port       string
	httpMethod string
	address    string
	conn       net.Conn
}

func NewHTTPClient(target, port string) *HTTPClient {
	return &HTTPClient{
		target:  target,
		port:    port,
		address: fmt.Sprintf("%s:%s", target, port),
	}
}

type Response struct {
	rawContent []byte
	_status    string
	_header    string
	_body      string
}

func NewResponse(buffer []byte) *Response {
	r := &Response{rawContent: buffer}
	r._extractMsg()
	return r
}

// status-line, header, bodyの値をフィールドに入れる。
func (resp *Response) _extractMsg() {
	resp.Status()
	resp.Header()
	resp.Body()
}

// HTTPステータスラインを取得する。すでに持っていればその値を返す。
func (resp *Response) Status() string {
	if resp._status != "" {
		return resp._status
	}

	// rawContentからstatusの内容を取得する。
	i := bytes.Index(resp.rawContent, []byte(crlf))
	resp._status = string(resp.rawContent[:i])

	return resp._status
}

// HTTPレスポンスヘッダーを取得する。すでに持っていればその値を返す。
func (resp *Response) Header() string {
	if resp._header != "" {
		return resp._header
	}

	// rawContentからheaderの内容を取得する。
	i := bytes.Index(resp.rawContent, []byte(crlf)) + len(crlf) // heaaderとstatusの境目
	j := bytes.Index(resp.rawContent, []byte(crlf+crlf))        // headerとbodyの境目
	resp._header = string(resp.rawContent[i:j])

	return resp._header
}

// HTTPレスポンスボディを取得する。すでに持っていればその値を返す。
func (resp *Response) Body() string {
	if resp._body != "" {
		return resp._body
	}

	// rawContentからbodyの内容を取得する。
	i := bytes.Index(resp.rawContent, []byte(crlf+crlf)) // headerとbodyの境目
	rawBody := resp.rawContent[i:]

	crlf := []byte("\r\n")
	rawBody = bytes.Trim(rawBody, string(crlf))
	j := bytes.Index(rawBody, crlf) + len(crlf) + 1 // chunke-sizeの境目

	k := bytes.LastIndex(rawBody, crlf) // 末尾の0との境目
	resp._body = string(rawBody[j:k])

	// TODO: chuke-sizeも抽出予定
	return resp._body
}

// HTTPレスポンスメッセージを受け取り、その内容をResponse構造体に含めて返す。
func (c HTTPClient) getHTTPResponse() (*Response, error) {

	err := c.sendHTTPRequest()
	if err != nil {
		return nil, err
	}

	rawResponse, err := c.readAllHTTPResponse()
	if err != nil {
		return nil, err
	}

	resp := NewResponse(rawResponse)
	return resp, err
}

// TCPでの接続を行う。
func (c *HTTPClient) _connect() error {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

// TCPでメッセージを送る。
func (c *HTTPClient) _write(msg []byte) error {
	_, err := c.conn.Write(msg)
	if err != nil {
		return err
	}

	return nil
}

// HTTP version 1.1
// HTTPリクエストを投げる。
func (c *HTTPClient) sendHTTPRequest() error {
	var (
		// とりあえずhttpメソッドは固定。今後変更予定。
		requestLine        = "GET / HTTP/1.1"
		requestHeader      = fmt.Sprintf("Host: %s \r\nConnection: close", c.target)
		httpRequestMessage = []byte(fmt.Sprint(requestLine, crlf, requestHeader, crlf, crlf))
	)

	if err := c._connect(); err != nil {
		fmt.Printf("can not connect to target(%s) error: %s \n", c.target, err)
		// TODO: どの箇所でエラーが生じたのかがわかるようにする。
		os.Exit(1)
	}

	err := c._write(httpRequestMessage)
	if err != nil {
		return err
	}

	return nil
}

// 受け取ったHTTPレスポンスメッセージの内容を全て読みとる。
func (c HTTPClient) readAllHTTPResponse() ([]byte, error) {
	c.conn.SetReadDeadline(time.Now().Add(time.Second * 5))

	slice := make([]byte, 1024)
	var buffer []byte
	var n int = 1

	for n != 0 {
		n, err := c.conn.Read(slice)
		if err != nil {
			if err == io.EOF {
				fmt.Println("done!")
				break
			}
			fmt.Println("can not read response error: ", err)
			return buffer, err
		}
		buffer = append(buffer, slice[:n]...)
	}

	return buffer, nil
}
