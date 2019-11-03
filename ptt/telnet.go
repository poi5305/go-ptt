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

// NewTelnet new Telnet
func NewTelnet(in, out chan byte, useStdin bool) *Telnet {
	c := &Telnet{
		inputChan:  in,
		outputChan: out,
		useStdin:   useStdin,
	}
	return c
}

// Telnet for go-telnet
type Telnet struct {
	telnetWriter telnet.Writer
	telnetReader telnet.Reader
	ctx          telnet.Context

	inputChan  chan byte
	outputChan chan byte
	useStdin   bool
}

func (c *Telnet) init() {
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

	if c.useStdin {
		c.dealWithStdinScanner()
	} else {
		c.dealWithInputChan()
	}
	time.Sleep(500 * time.Millisecond)
}

func (c *Telnet) dealWithStdinScanner() {
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
}

func (c *Telnet) dealWithInputChan() {
	for {
		b, ok := <-c.inputChan
		if !ok {
			break
		}
		err := oi.LongWriteByte(c.telnetWriter, b)
		if nil != err {
			break
		}
	}
}

// CallTELNET called by go-telnet for init
func (c *Telnet) CallTELNET(ctx telnet.Context, w telnet.Writer, r telnet.Reader) {
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

// func (c *Telnet) Write(bs []byte) (int, error) {
// 	fmt.Println("Write", len(bs))
// 	return len(bs), nil
// }

// func (c *Telnet) Read(bs []byte) (int, error) {
// 	fmt.Println("Read", len(bs))
// 	return len(bs), nil
// }
