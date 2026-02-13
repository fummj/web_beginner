package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

type AlternateBuffer struct {
	fd            int
	OldState      *term.State
	width, height int
	t             *term.Terminal
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

	return &AlternateBuffer{
		fd:       fd,
		OldState: OldState,
		width:    w,
		height:   h,
		t:        NewTerminal(w, h),
	}
}

type TerminalReadWriter struct {
	io.Reader
	io.Writer
}

func NewTerminalReadWriter() *TerminalReadWriter {
	return &TerminalReadWriter{os.Stdin, os.Stdout}
}

func NewTerminal(w, h int) *term.Terminal {
	rw := NewTerminalReadWriter()
	t := term.NewTerminal(rw, blink)

	t.SetSize(w, h)
	return t
}

func (ab AlternateBuffer) Enter() {

	ab.t.Write([]byte(enterESC))
	ab.t.Write([]byte(StrGreen))
	ab.t.Write([]byte(BgColor))
	ab.t.Write([]byte(Clear))

	ab.DrawTUI()

	defer ab.Restore()
}

func (ab AlternateBuffer) DrawTUI() {
	var (
		startVPoint int = ab.height / 4
		startHPoint int = int(float32(ab.width) / 3.3)
	)

	// wrttin.in/tokyoにcurlを飛ばした際に表示されるような、線にしたい。点線ではなく。やり方わからぬ
	fmt.Print("\x1b[", startVPoint+1, ";", startHPoint, "H", "+------------------------------------ HTTP REQUEST MESSAGE --------------------------------------+")
	fmt.Print("\x1b[", startVPoint+2, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+3, ";", startHPoint, "H", "+----------------------------------------- REQUEST LINE -----------------------------------------+")
	fmt.Print("\x1b[", startVPoint+4, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+5, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+6, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+7, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+8, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+9, ";", startHPoint, "H", "+---------------------------------------- REQUEST HEADER ----------------------------------------+")
	fmt.Print("\x1b[", startVPoint+10, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+11, ";", startHPoint, "H", "| Host:                                                                                          |")
	fmt.Print("\x1b[", startVPoint+12, ";", startHPoint, "H", "| Content-type:                                                                                  |")
	fmt.Print("\x1b[", startVPoint+13, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+14, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+15, ";", startHPoint, "H", "+----------------------------------------- REQUEST BODY -----------------------------------------+")
	fmt.Print("\x1b[", startVPoint+16, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+17, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+18, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+19, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+20, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+21, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+22, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+23, ";", startHPoint, "H", "|                                                                                                |")
	fmt.Print("\x1b[", startVPoint+24, ";", startHPoint, "H", "+------------------------------------------------------------------------------------------------+")
	fmt.Print("\x1b[", startVPoint+25, ";", startHPoint, "H", "                                                                                                  ")
	fmt.Print("\x1b[", startVPoint+26, ";", startHPoint, "H", "                             ++++++++++++                xxxxxxxxxxxx                             ")
	fmt.Print("\x1b[", startVPoint+27, ";", startHPoint, "H", "                             +   SEND   +                x  CANCEL  x                             ")
	fmt.Print("\x1b[", startVPoint+28, ";", startHPoint, "H", "                             ++++++++++++                xxxxxxxxxxxx                             ")

	ab.InputRequestContent(startVPoint, startHPoint)

	// time.Sleep(time.Second * 1 / 2)
}

func (ab AlternateBuffer) InputRequestContent(vPoint, hPoint int) {
	var (
		requestLineVPoint                  int = vPoint + 5
		requestLineAndHeaderHPoint         int = hPoint + 17
		requestHeaderHostVPoint            int = requestLineVPoint + 5
		requestHeaderContentTypeHostVPoint int = requestHeaderHostVPoint + 1
		requestBodyVPoint                  int = requestHeaderContentTypeHostVPoint + 5
		requestBodyHPoint                  int = hPoint + 1
	)

	ab._inputRequestContent(requestLineVPoint, requestLineAndHeaderHPoint)
	ab._inputRequestContent(requestHeaderHostVPoint, requestLineAndHeaderHPoint)
	ab._inputRequestContent(requestHeaderContentTypeHostVPoint, requestLineAndHeaderHPoint)
	ab._inputRequestContent(requestBodyVPoint, requestBodyHPoint)
}

func (ab AlternateBuffer) _inputRequestContent(vPoint, hPoint int) {
	// カーソル移動(初期位置)
	ab.t.Write([]byte(Home))

	for i := 0; i < vPoint; i++ {
		ab.t.Write([]byte(cursorDownESC))
	}

	for i := 0; i < hPoint; i++ {
		ab.t.Write([]byte(cursorRightESC))
	}

	_, err := ab.t.ReadLine()
	if err != nil {
		os.Exit(1)
	}
}

func (ab AlternateBuffer) verifyInputKey() {
	// TODO: 標準入力から1バイトずつキーを取得したい。
}

func (ab AlternateBuffer) Restore() {
	err := term.Restore(ab.fd, ab.OldState)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Stdout.Write([]byte(exitESC))
}
