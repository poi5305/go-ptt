package terminal

import (
	"fmt"

	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
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

	board [][]*Rune
	// decoder
	decoder     *encoding.Decoder
	runeFromBuf []byte
	runeToBuf   []byte
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
		r := t.decode(b)
		if r != utf8.RuneError {
			t.putRune(r)
		}
	}
}

func (t *Terminal) decode(b byte) rune {
	t.runeFromBuf = append(t.runeFromBuf, b)
	if t.decoder == nil {
		r, l := utf8.DecodeRune(t.runeFromBuf)
		if r != utf8.RuneError {
			t.runeFromBuf = t.runeFromBuf[l:]
		}
		if len(t.runeFromBuf) >= 4 {
			t.runeFromBuf = make([]byte, 0, 4)
			fmt.Println("decoder error bytes to utf8", t.runeFromBuf)
		}
		return r
	}
	nd, ns, err := t.decoder.Transform(t.runeToBuf, t.runeFromBuf, false)
	if err != nil && err != transform.ErrShortSrc {
		fmt.Println("decoder error", err, len(t.runeFromBuf))
		return utf8.RuneError
	}
	str := string(t.runeToBuf[:nd])
	runes := []rune(str)
	fmt.Println(t.runeToBuf, t.runeFromBuf, nd, ns, str, runes)
	t.runeFromBuf = t.runeFromBuf[ns:]
	if len(runes) != 1 {
		fmt.Println("decoder warning length != 1", err, len(runes))
	}
	return runes[0]
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
		t.lineFeed()
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

func (t *Terminal) clearAll() {
	board := make([][]*Rune, t.height, t.height)
	for i := range board {
		board[i] = make([]*Rune, t.width, t.width)
	}
	t.board = board
	t.x = 0
	t.y = 0
}

func (t *Terminal) clearLine(y int) {
	if !t.assertInBoard(0, y) {
		return
	}
	t.board[y] = make([]*Rune, t.width, t.width)
}

func (t *Terminal) backspace() {
	t.lineLeft(t.y, t.x, t.width, 1)
	t.goXY(t.x-1, t.y)
}

func (t *Terminal) lineFeed() {
	if t.y < t.width {
		t.y++
	} else {
		t.lineUp(0, t.height, 1)
	}
}

func (t *Terminal) carriageReturn() {
	t.x = 0
}

func (t *Terminal) putRune(r rune) {
	if !t.assertInBoard(t.x, t.y) {
		return
	}
	nx := t.x
	t.board[t.y][t.x] = &Rune{
		r: r,
	}
	if r >= 32 && r < 127 {
		// 1 char
		nx++
	} else if t.x+1 < t.width {
		// 2 char
		t.board[t.y][t.x+1] = nil
		nx += 2
	}
	if nx >= t.width {
		t.lineFeed()
		t.x = 0
	}
}

func (t *Terminal) tab() {
	mod := t.x % 4
	nx := t.x + (4 - mod)
	for x := t.x; x < nx; x++ {
		t.board[t.y][x] = nil
	}
	t.board[t.y][t.x] = &Rune{
		r: rune('\t'),
	}
	t.goXY(nx, t.y)
}

func (t *Terminal) goXY(x, y int) {
	if x >= t.width {
		x = t.width - 1
	}
	if y >= t.height {
		y = t.height - 1
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	t.x = x
	t.y = y
}

func (t *Terminal) lineLeft(y, from, to, n int) {
	if !t.assertInBoard(from-n, y) || !t.assertInBoard(to-1, y) {
		return
	}
	if n >= t.width {
		t.clearLine(y)
		return
	}
	for x := from; x < to; x++ {
		nx := x - n
		t.board[y][nx] = t.board[y][x]
	}
	for i := 0; i < n; i++ {
		x := to - i - 1
		t.board[y][x] = nil
	}
}

func (t *Terminal) lineRight(y, from, to, n int) {
	if !t.assertInBoard(from, y) || !t.assertInBoard(to-1+n, y) {
		return
	}
	if n >= t.width {
		t.clearLine(y)
		return
	}
	for x := to - 1; x >= from; x-- {
		nx := x + n
		t.board[y][nx] = t.board[y][x]
	}
	for i := 0; i < n; i++ {
		x := from + i
		t.board[y][x] = nil
	}
}

func (t *Terminal) lineUp(from, to, n int) {
	if !t.assertInBoard(0, from) || !t.assertInBoard(0, to-1) {
		return
	}
	if n >= t.height {
		t.clearAll()
		return
	}
	for y := from; y < to; y++ {
		ny := y - n
		if ny < 0 {
			continue
		}
		t.board[ny] = t.board[y]
	}
	for i := 0; i < n; i++ {
		y := to - i - 1
		t.clearLine(y)
	}
}

func (t *Terminal) lineDown(from, to, n int) {
	if !t.assertInBoard(0, from) || !t.assertInBoard(0, to-1) {
		return
	}
	if n >= t.height {
		t.clearAll()
		return
	}
	for y := to - 1; y <= from; y-- {
		ny := y + n
		if ny >= t.height {
			continue
		}
		t.board[ny] = t.board[y]
	}
	for i := 0; i < n; i++ {
		y := from + i
		t.clearLine(y)
	}
}

func (t *Terminal) assertInBoard(x, y int) bool {
	if x < 0 || x >= t.width {
		fmt.Println("Overflow assertInBoard x", x)
		return false
	}
	if y < 0 || y >= t.height {
		fmt.Println("Overflow assertInBoard y", y)
		return false
	}
	return true
}
