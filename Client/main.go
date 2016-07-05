package main

import (
	"fmt"
	"time"

	"VODone/Client/msgs"
	"bitbucket.org/serverFramework/serverFramework/utils"
)

const (
	defaultBufferSize = 16 * 1024
	ProtocolHeaderLen = 9 // header1 + cmd4 + length4

	C2LoginServerHB = 5 * time.Second
	C2QueueServerHB = 5 * time.Second
	C2QueueServerPP = 1 * time.Second
)

var wg utils.WaitGroupWrapper
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
	conn := connect2LoginServer(addr)
	sendLoginPakcet(conn, 12) // 10s以后发送登录包

	wg.Wait()

	fmt.Printf("client exit\n")
}
