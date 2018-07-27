package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	websocket2 "github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

func getSSHClient() (io.Reader, io.Writer) {
	config := &ssh.ClientConfig{
		User: "bbs",
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
	url := "wss://ws.ptt.cc/bbsu"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	return ws, ws
}

func getWebsocketClient2() (io.Reader, io.Writer) {
	origin := "https://www.ptt.cc"
	url := "wss://ws.ptt.cc/bbsu"
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

func main() {
	// r, w := getSSHClient()
	r, w := getWebsocketClient()

	// bs := make([]byte, 1024, 1024)
	// go func() {
	// 	for {
	// 		r.Read(bs)
	// 		fmt.Println(string(bs))
	// 	}
	// }()
	// w.Write([]byte{IAC, WILL, TTYPE})
	// w.Write([]byte{IAC, SB, TTYPE, 0, 86, 84, 49, 48, 48, IAC, SE})
	// w.Write([]byte{IAC, WILL, NAWS})
	// w.Write([]byte{IAC, SB, 0, 80, 0, 24, IAC, SE})
	// w.Write([]byte{IAC, DO, ECHO})
	// w.Write([]byte{IAC, DO, SGA})
	// w.Write([]byte{IAC, DONOT, BINARY})
	// time.Sleep(time.Second)
	// w.Write([]byte("guest\r"))
	// time.Sleep(time.Second)
	// w.Write([]byte("\r\n\n\n\n"))
	// time.Sleep(time.Second)

	// time.Sleep(time.Second)
	// time.Sleep(time.Second)

	vm := NewVM(r, w)
	time.Sleep(2 * time.Second)
	// vm.writeInitialMessage()
	fmt.Println("Guest")
	vm.writeGuest()
	time.Sleep(1 * time.Second)
	for {
		vm.printBoard()
		time.Sleep(time.Second * 3)
	}

}
