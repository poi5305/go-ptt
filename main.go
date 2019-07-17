package main

import (
	"fmt"
	"os"
	"time"

	"github.com/reiver/go-oi"
	"github.com/reiver/go-telnet"
)

func main() {
	var caller telnet.Caller = NewCaller()

	// var caller telnet.Caller = telnet.StandardCaller

	// io := &IO{}
	// context := telnet.NewContext()
	// caller.CallTELNET(context, io, io)

	telnet.DialToAndCall("ptt.cc:23", caller)
	select {}
}

func NewCaller() *Caller {
	c := &Caller{}
	return c
}

type Caller struct {
	writer telnet.Writer
	reader telnet.Reader
	ctx    telnet.Context
}

func (c *Caller) init() {
	go func() {
		p := make([]byte, 1, 1)
		for {
			n, err := c.reader.Read(p)
			if n <= 0 && nil == err {
				continue
			} else if n <= 0 && nil != err {
				break
			}
			oi.LongWrite(os.Stdout, p)
		}
	}()
	// very important, or connection will break
	time.Sleep(100 * time.Millisecond)
}

func (c *Caller) CallTELNET(ctx telnet.Context, w telnet.Writer, r telnet.Reader) {
	fmt.Println("CallTELNET")
	c.ctx = ctx
	c.writer = w
	c.reader = r
	c.init()
}

func (c *Caller) Write(bs []byte) (int, error) {
	fmt.Println("Write", len(bs))
	return len(bs), nil
}

func (c *Caller) Read(bs []byte) (int, error) {
	fmt.Println("Read", len(bs))
	return len(bs), nil
}
