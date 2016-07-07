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
	conn := Connect2LoginServer(addr)
	SendLoginPakcet(conn)

	msgs.WG.Wait()

	fmt.Printf("client exit\n")
}
