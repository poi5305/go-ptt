package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	websocket2 "github.com/gorilla/websocket"
	telnet "github.com/reiver/go-telnet"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/websocket"
)

type RW struct {
	R io.Reader
	W io.Writer
}

func (r *RW) Read(p []byte) (n int, err error) {
	return r.R.Read(p)
}

func (r *RW) Write(p []byte) (n int, err error) {
	// fmt.Println("Write", p)
	return len(p), nil
	// return r.W.Write(p)
}

func getSSHClient() (io.Reader, io.Writer) {
	config := &ssh.ClientConfig{
		User: "bbsu",
		Auth: []ssh.AuthMethod{
			ssh.Password(""),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", "ptt.cc:22", config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	session, err := client.NewSession()
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("VT100", 80, 40, modes); err != nil {
		session.Close()
		log.Fatalf("request for pseudo terminal failed: %s", err)
	}

	stdin, err := session.StdinPipe()
	stdout, err := session.StdoutPipe()

	go session.Start("")

	return stdout, stdin
}

func getWebsocketClient() (io.Reader, io.Writer) {
	origin := "https://www.ptt.cc"
	url := "wss://ws.ptt.cc/bbs"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	return ws, ws
}

func getWebsocketClient2() (io.Reader, io.Writer) {
	origin := "https://www.ptt.cc"
	url := "wss://ws.ptt.cc/bbs"
	hs := make(http.Header)
	hs["Origin"] = []string{origin}
	ws, _, err := websocket2.DefaultDialer.Dial(url, hs)
	if err != nil {
		log.Fatal(err)
	}
	m, r, _ := ws.NextReader()
	w, _ := ws.NextWriter(m)
	return r, w
}

func getTelnet() io.ReadWriter {
	conn, err := telnet.DialTo("ptt.cc:23")
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func main() {
	// var caller telnet.Caller = telnet.StandardCaller
	// rw := getTelnet()
	// r, w := getSSHClient()
	// r, w := getWebsocketClient()

	// r2, w1 := io.Pipe()
	r1, w2 := io.Pipe()

	rw := &RW{
		R: r1,
		W: nil,
	}

	term := terminal.NewTerminal(rw, "")

	term.SetSize(80, 24)
	// oldState, err := terminal.MakeRaw(0)
	// if err != nil {
	// 	panic(err)
	// }
	// defer terminal.Restore(0, oldState)

	origin := "https://www.ptt.cc"
	url := "wss://ws.ptt.cc/bbs"
	hs := make(http.Header)
	hs["Origin"] = []string{origin}
	ws, _, err := websocket2.DefaultDialer.Dial(url, hs)
	if err != nil {
		log.Fatal(err)
	}

	bs := make([]byte, 1024, 1024)

	fmt.Println(w2, term)
	// i, e := ws.Read(bs)
	// fmt.Println(i, e)

	go func() {
		for {
			t, r, _ := ws.NextReader()
			i, _ := r.Read(bs)
			s := string(bs[:i])
			// fmt.Print(s)
			w2.Write(bs[:i])
			// fmt.Println(ii)
			if strings.Contains(s, "new") {
				// t, r, _ = ws.NextReader()
				time.Sleep(time.Second)
				w, _ := ws.NextWriter(t)
				w.Write([]byte("guest\r\n"))
				w.Close()
				// term.Write([]byte("guest\r\n"))
				// fmt.Print("guest\r\n")
			}
		}
	}()

	// go func() {
	// 	bb := make([]byte, 1, 1)
	// 	for {
	// 		i, e := r2.Read(bb)
	// 		if e == nil {
	// 			fmt.Println(i, e, bs[:i])
	// 		}
	// 	}
	// }()

	go func() {
		for {
			line, err := term.ReadLine()
			if err == nil {
				fmt.Println(line)
			} else {
				fmt.Println(err)
				break
			}
		}
	}()

	time.Sleep(2 * time.Second)
	// t, _, _ := ws.NextReader()
	// w, _ := ws.NextWriter(t)
	// w.Write([]byte("guest\r\n"))
	// w.Close()
	// fmt.Print("guest\r\n")

	// ws.NextWriter(t)

	// w.Write([]byte{IAC, WILL, TTYPE})
	// w.Write([]byte{IAC, SB, TTYPE, 0, 86, 84, 49, 48, 48, IAC, SE})
	// w.Write([]byte{IAC, WILL, NAWS})
	// w.Write([]byte{IAC, SB, 0, 80, 0, 24, IAC, SE})
	// w.Write([]byte{IAC, DO, ECHO})
	// w.Write([]byte{IAC, DO, SGA})
	// w.Write([]byte{IAC, DONOT, BINARY})
	// time.Sleep(time.Second)
	// w, _ := ws.NewFrameWriter(ws.PayloadType)
	// w.Write([]byte("fqstory\r"))
	// w.Close()
	// time.Sleep(time.Second)
	// w.Write([]byte("\r\n\n\n\n"))
	// time.Sleep(time.Second)

	time.Sleep(time.Second)
	time.Sleep(time.Second)

	// vm := NewVM(r, w)
	// time.Sleep(2 * time.Second)
	// // vm.writeInitialMessage()
	// fmt.Println("Guest")
	// vm.writeGuest()
	// time.Sleep(1 * time.Second)
	// for {
	// 	vm.printBoard()
	// 	time.Sleep(time.Second * 3)
	// }

}
