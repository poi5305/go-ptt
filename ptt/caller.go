package ptt

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"time"

	oi "github.com/reiver/go-oi"
	telnet "github.com/reiver/go-telnet"
)

// NewCaller new Caller
func NewCaller(in chan byte, out chan byte) *Caller {
	c := &Caller{
		inputChan:  in,
		outputChan: out,
	}
	return c
}

// Caller for go-telnet
type Caller struct {
	telnetWriter telnet.Writer
	telnetReader telnet.Reader
	ctx          telnet.Context

	inputChan  chan byte
	outputChan chan byte
}

func (c *Caller) init() {
	go func() {
		p := make([]byte, 1, 1)
		for {
			n, err := c.telnetReader.Read(p)
			if n <= 0 && nil == err {
				continue
			} else if n <= 0 && nil != err {
				close(c.outputChan)
				close(c.inputChan)
				break
			}
			// oi.LongWrite(os.Stdout, p)
			c.outputChan <- p[0]
		}
	}()

	// go func() {
	// 	for {
	// 		b, ok := <-c.inputChan
	// 		if !ok {
	// 			break
	// 		}
	// 		oi.LongWriteByte(c.telnetWriter, b)
	// 	}
	// }()

	var buffer bytes.Buffer
	var p []byte

	var crlfBuffer [2]byte = [2]byte{'\r', '\n'}
	crlf := crlfBuffer[:]

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(scannerSplitFunc)

	for scanner.Scan() {
		buffer.Write(scanner.Bytes())
		buffer.Write(crlf)

		p = buffer.Bytes()

		n, err := oi.LongWrite(c.telnetWriter, p)
		if nil != err {
			break
		}
		if expected, actual := int64(len(p)), n; expected != actual {
			err := fmt.Errorf("Transmission problem: tried sending %d bytes, but actually only sent %d bytes.", expected, actual)
			fmt.Fprint(os.Stderr, err.Error())
			return
		}
		buffer.Reset()
	}

	time.Sleep(3 * time.Millisecond)
}

// CallTELNET called by go-telnet for init
func (c *Caller) CallTELNET(ctx telnet.Context, w telnet.Writer, r telnet.Reader) {
	fmt.Println("CallTELNET")
	c.ctx = ctx
	c.telnetWriter = w
	c.telnetReader = r
	c.init()
}

func scannerSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF {
		return 0, nil, nil
	}

	return bufio.ScanLines(data, atEOF)
}

// func (c *Caller) Write(bs []byte) (int, error) {
// 	fmt.Println("Write", len(bs))
// 	return len(bs), nil
// }

// func (c *Caller) Read(bs []byte) (int, error) {
// 	fmt.Println("Read", len(bs))
// 	return len(bs), nil
// }
