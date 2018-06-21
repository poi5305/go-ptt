package main

import (
	"bufio"
	"fmt"
	"io"
	"time"
)

type BState int

const (
	WROD = BState(0)
	CMD  = BState(1)

	ECHO  = byte(1)
	TTYPE = byte(24)
	ESC   = byte(27)
	NAWS  = byte(31)
	SE    = byte(240)
	SB    = byte(250)
	WILL  = byte(251)
	WONT  = byte(252)
	DO    = byte(253)
	DONOT = byte(254)
	IAC   = byte(255)
)

func NewVM(r io.Reader, w io.Writer) *VM {
	r1, w1 := io.Pipe()
	board := make([][]*Word, 24, 24)
	for i := range board {
		board[i] = make([]*Word, 80, 80)
	}
	vm := &VM{
		rawReader:  bufio.NewReader(r),
		rawWriter:  bufio.NewWriter(w),
		wordReader: bufio.NewReader(r1),
		wordWriter: bufio.NewWriter(w1),
		wordChan:   make(chan Word, 1),
		board:      board,
		asniState:  make(map[byte]string),
	}
	go vm.readLoop()
	return vm
}

type Word struct {
	w    rune
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

	board     [][]*Word
	cRow      int
	cCol      int
	bState    BState
	cmdBuf    []byte
	asniState map[byte]string
}

func (v *VM) writeInitialMessage() {
	v.rawWriter.Write([]byte{IAC, WILL, TTYPE})
	v.rawWriter.Write([]byte{IAC, SB, TTYPE, 0, 86, 84, 49, 48, 48, IAC, SE})
	v.rawWriter.Write([]byte{IAC, WILL, NAWS})
	v.rawWriter.Write([]byte{IAC, SB, 0, 80, 0, 24, IAC, SE})
	v.rawWriter.Write([]byte{IAC, DO, ECHO})
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
		bs, err = v.readNBytes(2)
		bs = append([]byte{IAC, buf[0]}, bs...)
	}
	return err
}

func (v *VM) onANSI() (string, error) {
	bs := make([]byte, 0, 8)
	buf := make([]byte, 1, 1)
	for {
		_, err := v.rawReader.Read(buf)
		if err != nil {
			return "", err
		}
		bs = append(bs, buf[0])
		switch buf[0] {
		case 'H', 'f', 'A', 'B', 'C', 'D', 's', 'u', 'J', 'K', 'm', 'h', 'l', 'p':
			return string(bs), nil
		}
	}
	return "", nil
}

func (v *VM) readLoop() {
	runeChan := make(chan rune, 4)
	go func() {
		for {
			r, _, err := v.wordReader.ReadRune()
			if err != nil {
				fmt.Println(err)
				break
			}
			runeChan <- r
		}
	}()

	buf := make([]byte, 1, 1)
	for {
		_, err := v.rawReader.Read(buf)
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
			cmd, err := v.onANSI()
			if err != nil {
				fmt.Println(err)
			}
			// fmt.Println("On ANSI", cmd)
			v.handleASNI(cmd)
			continue
		}
		v.wordWriter.WriteByte(b)
		v.wordWriter.Flush()

		select {
		case wordRune := <-runeChan:
			v.newWord(wordRune)
		case <-time.After(time.Microsecond):
		}
		// fmt.Println(v.wordWriter.Buffered())
	}
}

func (v *VM) handleASNI(cmd string) {
	asniType := cmd[len(cmd)-1]
	v.asniState[asniType] = cmd
	var param1 int
	var param2 int
	if asniType == 'H' || asniType == 'f' {
		fmt.Sscanf(cmd, "[%d;%d"+string(asniType), &param1, &param2) // row;col
		v.cRow = param1
		v.cCol = param2
	}
}

func (v *VM) newWord(w rune) {
	if w < 32 {
		fmt.Println(w)
		switch w {
		case 8:
			// v.board[v.cRow][v.cCol] = nil
			v.cCol--
		case 10:
			v.cRow++
			v.cCol = 0
		case 13:
			v.cCol = 0
		}
		return
	}

	n := len(string(w))
	ansi := v.asniState['m']
	fmt.Println(string(w), w, v.cRow, v.cCol, len(v.board))
	if v.cRow >= 24 || v.cCol >= 80 {
		fmt.Println("Error")
		return
	}

	v.board[v.cRow][v.cCol] = &Word{
		w:    w,
		raw:  v.cRow,
		col:  v.cCol,
		ansi: ansi,
	}
	v.cCol++
	if n > 1 {
		// v.board[v.cRow][v.cCol+1] = nil
		// v.cCol++
	}
}

func (v *VM) printBoard() {
	for row := range v.board {
		for col := range v.board[row] {
			word := v.board[row][col]
			if word != nil {
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
