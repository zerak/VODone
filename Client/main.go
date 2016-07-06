package main

import (
	"fmt"
	"time"

	"VODone/Client/msgs"
	. "VODone/Client/login"
	. "VODone/Client/queue"
)

var addr string
var Max, Cur, Time int32

func init() {
	MsgChanLogin = make(chan *msgs.Message, 10000)
	PoolLogin.New = func() interface{} {
		return &msgs.Message{
			Timestamp: time.Now(),
		}
	}
	ExitChanLogin = make(chan int, 1)

	MsgChanQueue = make(chan *msgs.Message, 10000)
	PoolQueue.New = func() interface{} {
		return &msgs.Message{
			Timestamp: time.Now(),
		}
	}
	ExitChanQueue = make(chan int, 1)

	addr = "127.0.0.1:60060"
}

func main() {
	// 1,first connect to login server
	conn := Connect2LoginServer(addr)
	SendLoginPakcet(conn, 3) // 10s以后发送登录包

	msgs.WG.Wait()

	fmt.Printf("client exit\n")
}
