package main

import (
	"bufio"
	"fmt"
	"io"

	"golang.org/x/text/encoding"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

type BState int

const (
	WROD = BState(0)
	CMD  = BState(1)

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

func NewVM(r io.Reader, w io.Writer) *VM {
	r1, w1 := io.Pipe()
	vm := &VM{
		rawReader:  bufio.NewReader(r),
		rawWriter:  bufio.NewWriter(w),
		wordReader: bufio.NewReader(r1),
		wordWriter: bufio.NewWriter(w1),
		wordChan:   make(chan Word, 1),
		runeChan:   make(chan rune, 1),
		byteChan:   make(chan byte, 1),
		asniState:  make(map[byte]string),
		decoder:    traditionalchinese.Big5.NewDecoder(),
	}
	vm.eraseBoard()
	go vm.readLoop()
	go vm.handleRunes()
	return vm
}

type Word struct {
	w    rune
	b    byte
	raw  int
	col  int
	ansi string
}

type VM struct {
	rawReader  *bufio.Reader
	rawWriter  *bufio.Writer
	wordReader *bufio.Reader
	wordWriter *bufio.Writer
	wordChan   chan Word
	runeChan   chan rune
	byteChan   chan byte

	board     [][]*Word
	cRow      int
	cCol      int
	bState    BState
	cmdBuf    []byte
	asniState map[byte]string
	decoder   *encoding.Decoder
}

func (v *VM) eraseBoard() {
	board := make([][]*Word, 24, 24)
	for i := range board {
		board[i] = make([]*Word, 80, 80)
	}
	v.board = board
	v.cRow = 0
	v.cCol = 0
}

func (v *VM) writeInitialMessage() {
	// v.rawWriter.Write([]byte{IAC, WILL, TTYPE})
	// v.rawWriter.Write([]byte{IAC, SB, TTYPE, 0, 86, 84, 49, 48, 48, IAC, SE})
	// v.rawWriter.Write([]byte{IAC, WILL, NAWS})
	// v.rawWriter.Write([]byte{IAC, SB, 0, 80, 0, 24, IAC, SE})
	// v.rawWriter.Write([]byte{IAC, DO, ECHO})
	// v.rawWriter.Flush()
}

func (v *VM) writeGuest() {
	cmd := []byte("\r")
	fmt.Println("writeGuest")
	for _, c := range cmd {
		v.byteChan <- c
	}
	v.rawWriter.Write(cmd)
	v.rawWriter.Flush()
}

func (v *VM) readNBytes(n int) ([]byte, error) {
	bs := make([]byte, 0, n)
	buf := make([]byte, 1, 1)
	for i := 0; i < n; i++ {
		_, err := v.rawReader.Read(buf)
		if err != nil {
			return nil, err
		}
		bs = append(bs, buf[0])
	}
	return bs, nil
}

func (v *VM) readUntil(b byte) ([]byte, error) {
	bs := make([]byte, 0, 4)
	buf := make([]byte, 1, 1)
	for {
		_, err := v.rawReader.Read(buf)
		if err != nil {
			return nil, err
		}
		bs = append(bs, buf[0])
		if buf[0] == b {
			break
		}
	}
	return bs, nil
}

func (v *VM) onCommand() error {
	buf := make([]byte, 1, 1)
	var bs []byte
	_, err := v.rawReader.Read(buf)
	if err != nil {
		return err
	}
	if buf[0] == SB {
		bs, err = v.readUntil(SE)
		bs = append([]byte{IAC, buf[0]}, bs...)
	} else {
		bs, err = v.readNBytes(1)
		bs = append([]byte{IAC, buf[0]}, bs...)
	}
	fmt.Println("onCommand", bs)
	return v.replyCommand(bs)
}

func (v *VM) replyCommand(cmd []byte) error {
	if cmd == nil || len(cmd) < 3 {
		return nil
	}
	if cmd[1] == DO && cmd[2] == TTYPE {
		v.rawWriter.Write([]byte{IAC, WILL, TTYPE})
	} else if cmd[1] == SB && cmd[2] == TTYPE {
		v.rawWriter.Write([]byte{IAC, SB, TTYPE, 0, 86, 84, 49, 48, 48, IAC, SE})
	} else if cmd[1] == DO && cmd[2] == NAWS {
		v.rawWriter.Write([]byte{IAC, WILL, NAWS})
		v.rawWriter.Write([]byte{IAC, SB, 0, 80, 0, 24, IAC, SE})
	} else if cmd[1] == WILL && cmd[2] == ECHO {
		v.rawWriter.Write([]byte{IAC, DO, ECHO})
	} else if cmd[1] == WILL && cmd[2] == SGA {
		v.rawWriter.Write([]byte{IAC, DO, SGA})
	} else if cmd[1] == WILL && cmd[2] == BINARY {
		v.rawWriter.Write([]byte{IAC, DONOT, BINARY})
	}
	return v.rawWriter.Flush()
}

func (v *VM) onANSI() string {
	ansi := make([]byte, 0, 8)
	for {
		r := <-v.byteChan
		ansi = append(ansi, r)
		switch r {
		case 'H', 'f', 'A', 'B', 'C', 'D', 's', 'u', 'J', 'K', 'm', 'h', 'l', 'p':
			return string(ansi)
		}
	}
}

func (v *VM) onBig5(runeUtf8Buf, runeBig5Buf []byte) (int, int) {
	nd, ns, err := v.decoder.Transform(runeUtf8Buf, runeBig5Buf, false)
	if err != nil && err != transform.ErrShortSrc {
		fmt.Println("big5 decoder", err, len(runeBig5Buf))
		return 0, 0
	}
	// str := string(runeUtf8Buf[:nd])

	// fmt.Println(str, runeUtf8Buf, runeBig5Buf, nd, ns)

	// runes := []rune(str)
	// for _, r := range runes {
	// v.runeChan <- r
	// }

	return nd, ns
}

func (v *VM) readLoop() {
	// runeChan := make(chan rune, 4)

	buf := make([]byte, 1, 1)
	// runeBig5Buf := make([]byte, 0, 2)
	// runeUtf8Buf := make([]byte, 4, 4)
	for {
		_, err := v.rawReader.Read(buf)
		fmt.Println("FFFF", buf)
		if err != nil {
			fmt.Println(err)
			break
		}
		b := buf[0]
		if b == IAC {
			err := v.onCommand()
			if err != nil {
				fmt.Println(err)
			}
			continue
		} else if b == ESC {
			// cmd := v.onANSI()
			// runes := []rune(cmd)
			// for _, r := range runes {
			// 	v.runeChan <- r
			// }
			// fmt.Println("On ANSI", cmd)
			// v.handleASNI(cmd)
			// continue
		}
		v.byteChan <- b
		// runeBig5Buf = append(runeBig5Buf, b)
		// if len(runeBig5Buf) == 2 {
		// 	// decode big5 -> utf8
		// 	ns := v.onBig5(runeUtf8Buf, runeBig5Buf)
		// 	runeBig5Buf = runeBig5Buf[ns:]
		// }
	}
}

func (v *VM) handleRunes() {
	runeBig5Buf := make([]byte, 0, 2)
	runeUtf8Buf := make([]byte, 4, 4)
	for {
		b := <-v.byteChan
		ansi := ""
		switch {
		case b == ESC:
			ansi = v.onANSI()
			v.handleASNI(ansi)
			continue
		case b < 32:
			v.handleCtrlByte(b)
			continue
		}

		word := &Word{
			b:    b,
			raw:  v.cRow,
			col:  v.cCol,
			ansi: ansi,
		}

		runeBig5Buf = append(runeBig5Buf, b)
		// if len(runeBig5Buf) == 2 {
		// decode big5 -> utf8
		nd, ns := v.onBig5(runeUtf8Buf, runeBig5Buf)

		str := string(runeUtf8Buf[:nd])
		runes := []rune(str)
		fmt.Println(runeUtf8Buf, runeBig5Buf, nd, ns, str, runes)
		runeBig5Buf = runeBig5Buf[ns:]
		if ns == 1 {
			word.w = runes[0]
		} else if ns == 2 {
			if v.cCol-1 >= 0 {
				v.board[v.cRow][v.cCol-1].w = runes[0]
			}
		} else {
			// word.w = rune("")
		}
		if v.cRow < 24 && v.cCol < 80 {
			v.board[v.cRow][v.cCol] = word
			v.cCol++
		} else {
			fmt.Println("row col", v.cRow, v.cCol)
		}

		// }

		// r := <-v.runeChan
		// ansi := ""
		// switch {
		// case r == rune(ESC):
		// 	ansi = v.onANSI()
		// 	v.handleASNI(ansi)
		// 	continue
		// case r < 32:
		// 	v.handleCtrlByte(r)
		// 	continue
		// }
		// n := len(string(r))
	}
}

func (v *VM) handleASNI(cmd string) {
	fmt.Println("handleASNI", cmd)
	asniType := cmd[len(cmd)-1]
	v.asniState[asniType] = cmd
	var param1 int
	var param2 int
	switch asniType {
	case 'H', 'f':
		fmt.Sscanf(cmd, "[%d;%d"+string(asniType), &param1, &param2) // row;col
		if param1 != 0 {
			param1--
		}
		if param2 != 0 {
			param2--
		}
		if param1 >= 24 {
			param1 = 23
		}
		if param2 >= 80 {
			param2 = 79
		}
		v.cRow = param1
		v.cCol = param2
	case 'J':
		if cmd[len(cmd)-2] == '2' {
			// Erase Display
			v.eraseBoard()
		}
	case 'K':
		// Erase Line Clears all characters from the cursor position to the end of the line (including the character at the cursor position).
		fmt.Println(v.cRow, v.cCol)
		for i := v.cCol; i < len(v.board[v.cRow]); i++ {
			v.board[v.cRow][i] = nil
		}
	case 'm':
	default:
		fmt.Println("Error Unhandle ASNI", cmd)
	}
}

func (v *VM) handleCtrlByte(r byte) {
	switch r {
	case 8:
		if v.cCol > 0 {
			copy(v.board[v.cRow][v.cCol-1:], v.board[v.cRow][v.cCol:])
			v.board[v.cRow][len(v.board[v.cRow])-1] = nil
			v.cCol--
		}
	case 10:
		v.cRow++
		v.cCol = 0
	case 13:
		v.cCol = 0
	case 0:
	default:
		fmt.Println("Error Unhandle CtrlByte", int(r))
	}
}

// func (v *VM) newWord(w rune) {
// 	if w < 32 {
// 		fmt.Println(w)
// 		switch w {
// 		case 8:
// 			// v.board[v.cRow][v.cCol] = nil
// 			v.cCol--
// 		case 10:
// 			v.cRow++
// 			v.cCol = 0
// 		case 13:
// 			v.cCol = 0
// 		}
// 		return
// 	}

// 	n := len(string(w))
// 	ansi := v.asniState['m']
// 	fmt.Println(string(w), w, v.cRow, v.cCol, len(v.board))
// 	if v.cRow >= 24 || v.cCol >= 80 {
// 		fmt.Println("Error")
// 		return
// 	}

// 	v.board[v.cRow][v.cCol] = &Word{
// 		w:    w,
// 		raw:  v.cRow,
// 		col:  v.cCol,
// 		ansi: ansi,
// 	}
// 	v.cCol++
// 	if n > 1 {
// 		// v.board[v.cRow][v.cCol+1] = nil
// 		// v.cCol++
// 	}
// }

func (v *VM) printBoard() {
	for row := range v.board {
		for col := range v.board[row] {
			word := v.board[row][col]
			if word != nil && word.w != 0 {
				fmt.Print(string(word.w))
				// fmt.Print(word.w)
			}
		}
		fmt.Print(string("\n"))
	}
}

func (v *VM) read() ([]byte, error) {
	// bs := make([]byte, 1, 1)
	// v.wordReader.Read(bs)
	// fmt.Println("?", bs)
	// r, n, _ := v.wordReader.ReadRune()
	// fmt.Println(r, n)
	return nil, nil
}
