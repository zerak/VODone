package main

import (
	"fmt"

	. "VODone/Client/login"
	"VODone/Client/msgs"
)

var addr string

func init() {
	addr = "127.0.0.1:60060"
}

func main() {
	// 1,first connect to login server
	loginchan := make(chan bool)
	conn := Connect2LoginServer(addr, loginchan)
	<-loginchan
	SendLoginPakcet(conn)

	msgs.WG.Wait()

	fmt.Printf("client exit\n")
}
