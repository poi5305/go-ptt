package ptt

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

// NewTranslatorB2U new translator
func NewTranslatorB2U(in chan byte, out chan byte) *TranslatorB2U {
	t := &TranslatorB2U{
		in:  in,
		out: out,
		bh:  0,
		// bl:  0,
	}

	return t
}

// TranslatorB2U Big5 <-> UTF8
type TranslatorB2U struct {
	in  chan byte
	out chan byte

	bh byte
	c  int
	// bl byte
}

// big5 81~FE

func (t *TranslatorB2U) init() {
	go func() {
		for {
			b, ok := <-t.in
			if !ok {
				break
			}
			t.newByte(b)
			t.c++
		}
	}()
}

func (t *TranslatorB2U) newByte(b byte) {
	if t.bh == 0 {
		if b >= 0x81 && b <= 0xFE {
			t.bh = b
		} else {
			t.out <- b
		}
	} else if b == 27 { // ESC ignore ANSI words
		t.out <- b
		for {
			ansiByte, ok := <-t.in
			if !ok {
				break
			}
			t.out <- ansiByte
			if ansiByte == 109 { // 109 means 'm'
				break
			}
		}
	} else {
		utf8Value, ok := t.testBig5(t.bh, b)
		if ok {
			s := []byte(string(rune(utf8Value)))
			for _, v := range s {
				t.out <- v
			}
			t.bh = 0
		} else {
			fmt.Printf("Can not translate %d %X %X\n", t.c, t.bh, b)
			t.out <- t.bh
			t.out <- b
			t.bh = 0
		}
	}
}

func (t *TranslatorB2U) testBig5(bh byte, bl byte) (int, bool) {
	big5Value := (int(bh) << 8) | int(bl)
	utf8Value, ok := B2UTable[big5Value]
	return utf8Value, ok
}

// Big5ToUTF8 convert BIG5 to UTF-8
func Big5ToUTF8(s []byte) ([]byte, error) {
	bufReader := bytes.NewReader(s)
	O := transform.NewReader(bufReader, traditionalchinese.Big5.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}
