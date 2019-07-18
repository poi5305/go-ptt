package ptt

import (
	"fmt"
	"os"

	oi "github.com/reiver/go-oi"
	telnet "github.com/reiver/go-telnet"
)

// NewController new controller
func NewController() *Controller {
	ri := make(chan byte)
	ro := make(chan byte)
	bui := make(chan byte)
	buo := make(chan byte)
	caller := NewCaller(ri, ro)
	b2u := NewTranslatorB2U(bui, buo)
	b2u.init()

	c := &Controller{
		caller:        caller,
		b2u:           b2u,
		rawInputChan:  ri,
		rawOutputChan: ro,
		b2uInputChan:  bui,
		b2uOutputChan: buo,
	}

	return c
}

// Controller handle flow
type Controller struct {
	caller        *Caller
	b2u           *TranslatorB2U
	rawInputChan  chan byte
	rawOutputChan chan byte
	b2uInputChan  chan byte
	b2uOutputChan chan byte
	conn          *telnet.Conn
}

// Start start connection
func (c *Controller) Start() {
	if c.conn != nil {
		c.Stop()
	}

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
			b := <-c.b2uOutputChan
			oi.LongWriteByte(os.Stdout, b)
		}
	}()
	// time.Sleep(time.Second)

	// bs := []byte("\n")
	// for _, v := range bs {
	// 	c.rawInputChan <- v
	// }

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
	}
}
