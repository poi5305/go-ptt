package terminal

import (
	"fmt"

	"golang.org/x/text/encoding"
)

const (
	// WROD = BState(0)
	// CMD  = BState(1)

	BINARY = byte(0)
	ECHO   = byte(1)
	SGA    = byte(3)
	TTYPE  = byte(24)
	ESC    = byte(27)
	NAWS   = byte(31)
	SE     = byte(240)
	SB     = byte(250)
	WILL   = byte(251)
	WONT   = byte(252)
	DO     = byte(253)
	DONOT  = byte(254)
	IAC    = byte(255)
)

type Rune struct {
	r rune
}

type Terminal struct {
	IntputChan chan byte

	width  int
	height int
	x      int
	y      int

	board   [][]*Rune
	decoder *encoding.Decoder
}

func (t *Terminal) mainHandler() {
	for {
		b := <-t.IntputChan
		ansi := ""
		switch {
		case b == ESC:
			ansi = t.receiveANSI()
			t.handleASNI(ansi)
			continue
		case b < 32:
			t.handleCtrlByte(b)
			continue
		}
	}
}

func (t *Terminal) receiveANSI() string {
	ansi := make([]byte, 0, 8)
	for {
		r := <-t.IntputChan
		ansi = append(ansi, r)
		switch r {
		case 'H', 'f', 'A', 'B', 'C', 'D', 's', 'u', 'J', 'K', 'm', 'h', 'l', 'p':
			return string(ansi)
		}
	}
}

func (t *Terminal) handleASNI(cmd string) {
	fmt.Println("handleASNI", cmd)
	asniType := cmd[len(cmd)-1]
	// t.asniState[asniType] = cmd
	var param1 int
	var param2 int
	switch asniType {
	case 'H', 'f':
		fmt.Sscanf(cmd, "[%d;%d"+string(asniType), &param1, &param2) // row;col
		// if param1 != 0 {
		// 	param1--
		// }
		// if param2 != 0 {
		// 	param2--
		// }
		// if param1 >= 24 {
		// 	param1 = 23
		// }
		// if param2 >= 80 {
		// 	param2 = 79
		// }
		// v.cRow = param1
		// v.cCol = param2
	case 'J':
		if cmd[len(cmd)-2] == '2' {
			// Erase Display
			// v.eraseBoard()
		}
	case 'K':
		// Erase Line Clears all characters from the cursor position to the end of the line (including the character at the cursor position).
		// fmt.Println(v.cRow, v.cCol)
		// for i := v.cCol; i < len(v.board[v.cRow]); i++ {
		// 	v.board[v.cRow][i] = nil
		// }
	case 'm':
	default:
		fmt.Println("Error Unhandle ASNI", cmd)
	}
}

func (t *Terminal) handleCtrlByte(r byte) {
	switch r {
	case '\x07':
		break
	case '\b':
		t.backspace()
		// if v.cCol > 0 {
		// 	copy(v.board[v.cRow][v.cCol-1:], v.board[v.cRow][v.cCol:])
		// 	v.board[v.cRow][len(v.board[v.cRow])-1] = nil
		// 	v.cCol--
		// }
	case '\n', '\f', 'v':
		t.newLine()
		// v.cRow++
		// v.cCol = 0
	case '\r':
		t.carriageReturn()
		// v.cCol = 0
	case '\t':
	case '\x00':
	default:
		fmt.Println("Error Unhandle CtrlByte", int(r))
	}
}

func (t *Terminal) backspace() {

}

func (t *Terminal) newLine() {

}

func (t *Terminal) carriageReturn() {

}
