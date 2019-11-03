package ptt

import (
	"fmt"
	"strings"
	"time"

	telnet "github.com/reiver/go-telnet"
)

// NewController new controller
func NewController() *Controller {
	c := &Controller{}

	return c
}

// Controller handle flow
type Controller struct {
	telnet        *Telnet
	b2u           *TranslatorB2U
	terminal      *Terminal
	rawInputChan  chan byte
	rawOutputChan chan byte
	b2uInputChan  chan byte
	b2uOutputChan chan rune
	terminalChan  chan rune
	conn          *telnet.Conn
}

// Start start connection
func (c *Controller) Start() {
	if c.conn != nil {
		c.Stop()
	}
	c.rawInputChan = make(chan byte)
	c.rawOutputChan = make(chan byte)
	c.b2uInputChan = make(chan byte)
	c.b2uOutputChan = make(chan rune)
	c.terminalChan = make(chan rune)
	c.telnet = NewTelnet(c.rawInputChan, c.rawOutputChan, false)
	c.b2u = NewTranslatorB2U(c.b2uInputChan, c.b2uOutputChan)
	c.b2u.init()
	c.terminal = NewTerminal(c.terminalChan)

	go func() {
		for {
			b, ok := <-c.rawOutputChan
			if !ok {
				break
			}
			c.b2uInputChan <- b
		}
	}()

	go func() {
		for {
			b, ok := <-c.b2uOutputChan
			if !ok {
				break
			}
			c.terminalChan <- b
		}
	}()

	go c.dial()

	time.Sleep(100 * time.Millisecond)
}

func (c *Controller) dial() {
	var err error
	c.conn, err = telnet.DialTo("ptt.cc:23")
	if nil != err {
		fmt.Println(err)
		return
	}
	client := &telnet.Client{Caller: c.telnet}
	client.Call(c.conn)
}

// Stop stop connection
func (c *Controller) Stop() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		close(c.b2uInputChan)
		close(c.b2uOutputChan)
	}
}

// WriteString send message out
func (c *Controller) WriteString(str string) {
	bs := []byte(str)
	for _, b := range bs {
		c.rawInputChan <- b
	}
}

// ReadBoard return current terminal buffer
func (c *Controller) ReadBoard() string {
	return c.terminal.GetBoardText(false)
}

// WaitUntilString wait board contain string, timeout in milliseconds
func (c *Controller) WaitUntilString(str string, timeout int64) bool {
	now := time.Now()
	for time.Now().Sub(now).Nanoseconds()/1000000 < timeout {
		board := c.terminal.GetBoardText(false)
		if strings.Contains(board, str) {
			fmt.Println(board)
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}
