package ptt

import (
	"fmt"

	telnet "github.com/reiver/go-telnet"
)

// NewController new controller
func NewController() *Controller {
	c := &Controller{}

	return c
}

// Controller handle flow
type Controller struct {
	caller        *Caller
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
	c.caller = NewCaller(c.rawInputChan, c.rawOutputChan)
	c.b2u = NewTranslatorB2U(c.b2uInputChan, c.b2uOutputChan)
	c.b2u.init()
	c.terminal = NewTerminal(c.terminalChan)

	go func() {
		for {
			b, ok := <-c.rawOutputChan
			if !ok {
				break
			}
			// oi.LongWriteByte(os.Stdout, b)
			c.b2uInputChan <- b
		}
	}()

	go func() {
		for {
			b, ok := <-c.b2uOutputChan
			if !ok {
				break
			}
			// s := []byte(string(rune(b)))
			// for _, v := range s {
			// 	oi.LongWriteByte(os.Stdout, v)
			// }
			c.terminalChan <- b
		}
	}()
	// time.Sleep(time.Second)

	// bs := []byte("\n")
	// for _, v := range bs {
	// 	c.rawInputChan <- v
	// }

	c.dial()
}

func (c *Controller) dial() {
	var err error
	c.conn, err = telnet.DialTo("ptt.cc:23")
	if nil != err {
		fmt.Println(err)
		return
	}
	client := &telnet.Client{Caller: c.caller}
	client.Call(c.conn)
}

// Stop stop connection
func (c *Controller) Stop() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		close(c.rawInputChan)
		close(c.rawOutputChan)
		close(c.b2uInputChan)
		close(c.b2uOutputChan)
	}
}
