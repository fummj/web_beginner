package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

const (
	EnterESC = "\x1b[?1049h"
	ExitESC  = "\x1b[?1049l"

	StrRed  = "\x1b[31m"
	BgBlack = "\x1b[48;2;0;0;0m"
	Home    = "\x1b[H"
	Clear   = "\x1b[H\x1b[2J"

	Blink = "\x1b[5m"

	CursorDownESC  = "\x1b[1B"
	CursorRightESC = "\x1b[1C"

	Tab       uint8 = 9
	Enter     uint8 = 13
	Backspace uint8 = 127
	CtrlH     uint8 = 8
	CtrlU     uint8 = 21
)

type RequestContent struct {
	requestLine              string
	requestHeaderHost        string
	requestHeaderContentType string
	requestBody              string
}

type AlternateBuffer struct {
	fd                            int
	OldState                      *term.State
	width, height, vPoint, hPoint int
	t                             *term.Terminal
	rw                            *TerminalReadWriter
	tuiText                       string
	tabCount                      int
	rc                            *RequestContent
	scs                           bool
}

func NewAlternateBuffer() *AlternateBuffer {
	fd := int(os.Stdin.Fd())
	OldState, err := term.GetState(fd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = term.MakeRaw(fd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	w, h, err := term.GetSize(fd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	t, rw := NewTerminal(w, h)

	return &AlternateBuffer{
		fd:       fd,
		OldState: OldState,
		width:    w,
		height:   h,
		t:        t,
		rw:       rw,
		rc:       &RequestContent{},
		scs:      true,
	}
}

type TerminalReadWriter struct {
	io.Reader
	io.Writer
}

func NewTerminalReadWriter() *TerminalReadWriter {
	return &TerminalReadWriter{os.Stdin, os.Stdout}
}

func NewTerminal(w, h int) (*term.Terminal, *TerminalReadWriter) {
	rw := NewTerminalReadWriter()
	t := term.NewTerminal(rw, Blink)

	t.SetSize(w, h)
	return t, rw
}

func (ab *AlternateBuffer) Enter() {

	ab.t.Write([]byte(EnterESC))
	ab.t.Write([]byte(StrRed))
	ab.t.Write([]byte(BgBlack))
	ab.t.Write([]byte(Clear))

	ab.DrawTUI()

	defer ab.Restore()
}

func (ab *AlternateBuffer) DrawTUI() {
	ab.vPoint = ab.height / 4
	ab.hPoint = int(float32(ab.width) / 3.3)

	ab.tuiText = fmt.Sprint(
		"\x1b[", ab.vPoint+1, ";", ab.hPoint, "H", "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ HTTP REQUEST MESSAGE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓",
		"\x1b[", ab.vPoint+2, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+3, ";", ab.hPoint, "H", "┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ REQUEST LINE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫",
		"\x1b[", ab.vPoint+4, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+5, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+6, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+7, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+8, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+6, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+7, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+8, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+9, ";", ab.hPoint, "H", "┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ REQUEST HEADER ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫",
		"\x1b[", ab.vPoint+10, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+11, ";", ab.hPoint, "H", "┃ Host:                                                                                          ┃",
		"\x1b[", ab.vPoint+12, ";", ab.hPoint, "H", "┃ Content-type:                                                                                  ┃",
		"\x1b[", ab.vPoint+13, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+14, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+15, ";", ab.hPoint, "H", "┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ REQUEST BODY ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫",
		"\x1b[", ab.vPoint+16, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+17, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+18, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+19, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+20, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+21, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+22, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+23, ";", ab.hPoint, "H", "┃                                                                                                ┃",
		"\x1b[", ab.vPoint+24, ";", ab.hPoint, "H", "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛",
		"\x1b[", ab.vPoint+25, ";", ab.hPoint, "H", "                                                                                                  ",
		"\x1b[", ab.vPoint+26, ";", ab.hPoint, "H", "                             ++++++++++++                xxxxxxxxxxxx                             ",
		"\x1b[", ab.vPoint+27, ";", ab.hPoint, "H", "                             +   SEND   +                x  CANCEL  x                             ",
		"\x1b[", ab.vPoint+28, ";", ab.hPoint, "H", "                             ++++++++++++                xxxxxxxxxxxx                             ",
	)
	fmt.Print(ab.tuiText)

	ab.InputRequestContent(ab.vPoint, ab.hPoint)
	ab.ReadEnter()
}

func (ab AlternateBuffer) RenderingRequestLine() {
	ab.moveCursorRequestLine()
	fmt.Print(ab.rc.requestLine)
}

func (ab AlternateBuffer) RenderingRequestHeaderHost() {
	ab.moveCursorRequestHeaderHost()
	fmt.Print(ab.rc.requestHeaderHost)
}

func (ab AlternateBuffer) RenderingRequestHeaderContentType() {
	ab.moveCursorRequestHeaderContentType()
	fmt.Print(ab.rc.requestHeaderContentType)
}

func (ab AlternateBuffer) RenderingRequestBody() {
	ab.moveCursorRequestBody()
	fmt.Print(ab.rc.requestBody)
}

func (ab *AlternateBuffer) InputRequestContent(vPoint, hPoint int) {
	ab.InputRequestLine()
	ab.InputRequestHeaderHost()
	ab.InputRequestHeaderContentType()
	ab.InputRequestBody()
}

func (ab *AlternateBuffer) InputRequestLine() {
	ab.moveCursorRequestLine()
	ab.ReadLine()
}

func (ab *AlternateBuffer) InputRequestHeaderHost() {
	ab.moveCursorRequestHeaderHost()
	ab.ReadLine()
}

func (ab *AlternateBuffer) InputRequestHeaderContentType() {
	ab.moveCursorRequestHeaderContentType()
	ab.ReadLine()
}

func (ab *AlternateBuffer) InputRequestBody() {
	ab.moveCursorRequestBody()
	ab.ReadLine()
}

func (ab AlternateBuffer) MoveCursorRequestContent() {
	if ab.tabCount == 0 {
		ab.moveCursorRequestLine()
	}

	if ab.tabCount == 1 {
		ab.RenderingRequestLine()
		ab.moveCursorRequestHeaderHost()
	}

	if ab.tabCount == 2 {
		ab.RenderingRequestLine()
		ab.RenderingRequestHeaderHost()
		ab.moveCursorRequestHeaderContentType()
	}

	if ab.tabCount == 3 {
		ab.RenderingRequestLine()
		ab.RenderingRequestHeaderHost()
		ab.RenderingRequestHeaderContentType()
		ab.moveCursorRequestBody()
	}
}

func (ab AlternateBuffer) moveCursorRequestLine() {
	v := ab.vPoint + 5
	h := ab.hPoint + 17

	ab.moveCursor(v, h)
}

func (ab AlternateBuffer) moveCursorRequestHeaderHost() {
	v := ab.vPoint + 10
	h := ab.hPoint + 17

	ab.moveCursor(v, h)
}

func (ab AlternateBuffer) moveCursorRequestHeaderContentType() {
	v := ab.vPoint + 11
	h := ab.hPoint + 17

	ab.moveCursor(v, h)
}

func (ab AlternateBuffer) moveCursorRequestBody() {
	v := ab.vPoint + 16
	h := ab.hPoint + 1

	ab.moveCursor(v, h)
}

func (ab AlternateBuffer) moveCursorSaveOrCancel() {
	if ab.scs == true {
		ab._moveCursorSave()
	} else {
		ab._moveCursorCancel()
	}
}

func (ab AlternateBuffer) _moveCursorSave() {
	v := ab.vPoint + 26
	h := ab.hPoint + 32

	ab.moveCursor(v, h)
}

func (ab AlternateBuffer) _moveCursorCancel() {
	v := ab.vPoint + 26
	h := ab.hPoint + 59

	ab.moveCursor(v, h)
}

func (ab AlternateBuffer) moveCursor(vPoint, hPoint int) {
	ab.t.Write([]byte(Home))

	for i := 0; i < vPoint; i++ {
		ab.t.Write([]byte(CursorDownESC))
	}

	for i := 0; i < hPoint; i++ {
		ab.t.Write([]byte(CursorRightESC))
	}
}

func (ab AlternateBuffer) _inputRequestField(buffer *[]byte) {
	if ab.tabCount == 0 {
		ab.rc.requestLine = string(*buffer)
	}

	if ab.tabCount == 1 {
		ab.rc.requestHeaderHost = string(*buffer)
	}

	if ab.tabCount == 2 {
		ab.rc.requestHeaderContentType = string(*buffer)
	}

	if ab.tabCount == 3 {
		ab.rc.requestBody = string(*buffer)
	}
}

func (ab AlternateBuffer) RenderingRequestContent(esc uint8, buffer *[]byte) {
	ab.MoveCursorRequestContent()
	ab._renderingBuffer(esc, buffer)

	if ab.tabCount == 0 {
		ab.rc.requestLine = string(*buffer)
		ab.RenderingRequestLine()
	}

	if ab.tabCount == 1 {
		ab.rc.requestHeaderHost = string(*buffer)
		ab.RenderingRequestLine()
		ab.RenderingRequestHeaderHost()
	}

	if ab.tabCount == 2 {
		ab.rc.requestHeaderContentType = string(*buffer)
		ab.RenderingRequestLine()
		ab.RenderingRequestHeaderHost()
		ab.RenderingRequestHeaderContentType()
	}

	if ab.tabCount == 3 {
		ab.rc.requestBody = string(*buffer)
		ab.RenderingRequestLine()
		ab.RenderingRequestHeaderHost()
		ab.RenderingRequestHeaderContentType()
		ab.RenderingRequestBody()
	}
}

func (ab *AlternateBuffer) _renderingBuffer(esc uint8, buffer *[]byte) {
	if esc == Backspace || esc == CtrlH || esc == CtrlU {
		ab.rw.remover(esc, buffer)
		ab._inputRequestField(buffer)
	} else {
		ab.rw.adder(esc, buffer)
		ab._inputRequestField(buffer)
	}
}

// TUIテキストを再レンダリングする
func (ab AlternateBuffer) _reRenderingTUIText() {
	fmt.Print(Clear)
	fmt.Print(ab.tuiText)

}

func (ab *AlternateBuffer) ReadEnter() {
	ab.moveCursorSaveOrCancel()

	r := make([]byte, 1)
	for {
		i, err := ab.rw.Read(r)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if i == 0 {
			os.Exit(1)
		}

		esc := r[0]
		if esc == Tab {
			ab.scs = !ab.scs
			ab.moveCursorSaveOrCancel()
		}

		if esc == Enter {
			break
		}
	}
}

func (ab *AlternateBuffer) ReadLine() {
	bff := make([]byte, 1)
	r := make([]byte, 1)

	for {
		i, err := ab.rw.Read(r)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		esc := r[0]
		if i == 0 || esc == Tab {
			if esc == Tab {
				ab.tabCount += 1
				break
			}
		}

		if esc == Enter {
			continue
		}

		ab._reRenderingTUIText()
		ab.RenderingRequestContent(esc, &bff)
	}
}

func (rw TerminalReadWriter) remover(esc uint8, buffer *[]byte) {
	l := len(*buffer)

	if l == 0 {
		return
	}

	if esc == Backspace || esc == CtrlH {
		rw._remover(buffer)
		return
	}

	if esc == CtrlU {
		for i := 0; i < l; i++ {
			rw._remover(buffer)
		}
		return
	}
}

func (rw TerminalReadWriter) _remover(buffer *[]byte) {
	*buffer = (*buffer)[:len(*buffer)-1]
}

func (rw TerminalReadWriter) adder(esc uint8, buffer *[]byte) {
	*buffer = append(*buffer, esc)
}

func (ab AlternateBuffer) Restore() {
	err := term.Restore(ab.fd, ab.OldState)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Stdout.Write([]byte(ExitESC))
}
