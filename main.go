package main

import (
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {

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

	// bs := make([]byte, 1, 1)
	// for {
	// 	n, err := reader.Read(bs)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		break
	// 	}
	// 	fmt.Println(n, bs)
	// }
	// session.Wait()
	vm := NewVM(stdout, stdin)

	vm.writeInitialMessage()
	time.Sleep(5 * time.Second)
	vm.printBoard()
	for {
		// fmt.Println("want read")
		vm.read()

	}

}
