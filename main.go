package main

import (
	"github.com/poi5305/go-ptt/ptt"
)

func main() {

	// i := 0xA950 // 43344  169 80
	// bs := make([]byte, 4)
	// binary.BigEndian.PutUint32(bs, uint32(i))
	// fmt.Println(bs)
	// s := ptt.B2UTable[i]
	// binary.BigEndian.PutUint32(bs, uint32(s))
	// fmt.Println(bs)
	// println(s, string(rune(s)), len([]byte(string(rune(s)))))

	controller := ptt.NewController()
	controller.Start()
	select {}
}
