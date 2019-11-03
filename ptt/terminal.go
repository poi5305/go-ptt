package ptt

import (
	"fmt"
)

// Rune contains rune, position, colors...
type Rune struct {
	r rune
	w int
}

// NewTerminal new terminal
func NewTerminal(in chan rune) *Terminal {
	t := &Terminal{
		in:     in,
		width:  80,
		height: 24,
	}
	t.init()
	return t
}

// Terminal vt100
type Terminal struct {
	in chan rune

	width  int
	height int
	x      int
	y      int

	board [][]*Rune
}

func (t *Terminal) init() {
	t.clearAll()
	go t.initParser()

	// go func() {
	// 	for {
	// 		time.Sleep(3 * time.Second)
	// 		text := t.GetBoardText(true)
	// 		fmt.Print(text)
	// 	}
	// }()
}

func (t *Terminal) initParser() {
	for {
		r, ok := <-t.in
		if !ok {
			break
		}
		t.parseInput(r)
	}
}

func (t *Terminal) parseInput(r rune) {
	switch {
	case r == 27: // ESC
		ansi := t.readANSI()
		t.commandANSI(ansi)
	case r < 32:
		// fmt.Print(string(r))
		t.commandCtrl(r)
	default:
		// fmt.Print(string(r))
		t.putRune(r)
	}
}

func (t *Terminal) readANSI() string {
	ansi := ""
	for {
		r := <-t.in
		ansi += string(r)
		switch r {
		case 'H', 'f', 'A', 'B', 'C', 'D', 's', 'u', 'J', 'K', 'm', 'h', 'l', 'p':
			return ansi
		}
	}
}

func (t *Terminal) commandCtrl(r rune) {
	switch r {
	case '\x07':
		break
	case '\b':
		// fmt.Print("CtrlBack")
		t.backspace()
	case '\n', '\f', 'v':
		// fmt.Print("<CN>")
		t.lineFeed()
	case '\r':
		// fmt.Print("<CR>")
		t.carriageReturn()
	case '\t':
	case '\x00':
	default:
		fmt.Println("Error Unhandle CtrlByte", int(r))
	}
}

func (t *Terminal) commandANSI(cmd string) {
	cmdType := cmd[len(cmd)-1]
	var param1 int
	var param2 int
	switch cmdType {
	case 'H', 'f':
		// fmt.Println(cmd)
		fmt.Sscanf(cmd, "[%d;%d"+string(cmdType), &param1, &param2) // row;col
		t.goXY(param2-1, param1-1)
	case 'J':
		// fmt.Println(cmd)
		if cmd[len(cmd)-2] == '2' {
			t.clearAll()
		}
	case 'K':
		// fmt.Println(cmd, t.y, t.x, t.width)
		t.clearLine(t.y, t.x, t.width)
	case 'm':
	default:
		fmt.Println("Error Unhandle ASNI", cmd)
	}
}

func (t *Terminal) clearAll() {
	board := make([][]*Rune, t.height, t.height)
	for i := range board {
		board[i] = make([]*Rune, t.width, t.width)
	}
	t.board = board
	// t.x = 0
	// t.y = 0
}

func (t *Terminal) goXY(x, y int) {
	// fmt.Printf("<GTA,%d,%d>", x, y)
	if x >= t.width {
		// x = t.width - 1
	}
	if y >= t.height {
		// y = t.height - 1
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

func (t *Terminal) assertInBoard(x, y int) bool {
	if x < 0 || x >= t.width {
		fmt.Println("Overflow assertInBoard x y", x, y)
		return false
	}
	if y < 0 || y >= t.height {
		fmt.Println("Overflow assertInBoard x y", x, y)
		return false
	}
	return true
}

func (t *Terminal) backspace() {
	if t.x > 0 {
		t.x--
	}
	// t.lineLeft(t.y, t.x, t.width, 1)
	// t.goXY(t.x-1, t.y)
}

func (t *Terminal) clearLine(y, from, to int) {
	if !t.assertInBoard(from, y) || !t.assertInBoard(to-1, y) {
		return
	}
	if from == 0 && to == t.width {
		t.board[y] = make([]*Rune, t.width, t.width)
		return
	}
	for x := from; x < to; x++ {
		t.board[y][x] = nil
	}
}

func (t *Terminal) putRune(r rune) {
	// if t.x >= t.width {
	// 	t.lineFeed()
	// 	t.x = 0
	// }
	// if !t.assertInBoard(t.x, t.y) {
	// 	fmt.Println("Exceed", string(r))
	// 	return
	// }
	t.board[t.y][t.x] = &Rune{
		r: r,
	}
	if r >= 32 && r < 127 {
		// 1 char
		t.board[t.y][t.x].w = 1
		t.x++
	} else if t.x+1 < t.width {
		// 2 char
		t.board[t.y][t.x].w = 2
		t.board[t.y][t.x+1] = nil
		t.x += 2
	}
	if t.x > t.width {
		// t.lineFeed()
		// fmt.Println("putRune lineFeed", t.x, t.y)
	}
}

// GetBoardText get current board text
func (t *Terminal) GetBoardText(index bool) string {
	text := ""
	if index {
		text = "\n===01234567890123456789012345678901234567890123456789012345678901234567890123456789\n"
	}
	for r, row := range t.board {
		text += fmt.Sprintf("%2d ", r)
		for x := 0; x < len(row); x++ {
			col := row[x]
			if col == nil {
				text += " "
				continue
			}
			if col.w == 2 {
				x++
			}
			text += string(col.r)
		}
		text += "\n"
	}
	if index {
		text += "===01234567890123456789012345678901234567890123456789012345678901234567890123456789\n"
	}
	return text
}

// lineFeed '\n'
func (t *Terminal) lineFeed() {
	// fmt.Printf("<lf>")
	if t.y < t.height-1 {
		t.y++
	} else {
		t.lineUp(0, t.height, 1)
	}
}

// carriageReturn '\r'
func (t *Terminal) carriageReturn() {
	t.x = 0
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
		t.clearLine(y, 0, t.width)
	}
}
