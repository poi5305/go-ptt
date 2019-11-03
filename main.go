package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/poi5305/go-ptt/ptt"
)

func main() {
	step := login("", "", 5000)
	fmt.Println(step)
}

// step < 3 failed, step = -1 password error, step == 3 user online, step == 4 success
func login(account, password string, timeout int64) int {
	if account == "" || password == "" {
		return -1
	}

	controller := ptt.NewController()
	controller.Start()
	now := time.Now()

	step := 0
	for time.Now().Sub(now).Nanoseconds()/1000000 < timeout {
		board := controller.ReadBoard()
		if step == 0 && strings.Contains(board, "請輸入代號") {
			step++
			controller.WriteString(account + "\r")
		} else if step == 1 && strings.Contains(board, "請輸入您的密碼") {
			step++
			controller.WriteString(password + "\r")
		} else if step == 2 && strings.Contains(board, "歡迎您再度拜訪") {
			step++
			controller.WriteString("\n")
		} else if step == 3 && strings.Contains(board, "您要刪除以上錯誤嘗試的記錄") {
			controller.WriteString("Y\r\n")
		} else if step == 2 && strings.Contains(board, "您想刪除其他重複登入的連線嗎") {
			step++
			controller.WriteString("N\r\n")
			break
		} else if step == 2 && strings.Contains(board, "密碼不對或無此帳號") {
			step = -1
			break
		} else if step == 3 && strings.Contains(board, "離開，再見") {
			fmt.Println("Goodbye")
			step++
			controller.WriteString("G\rY\r")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	controller.Stop()
	return step
}
