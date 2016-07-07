package msgs

import (
	. "bitbucket.org/serverFramework/serverFramework/core"
	"strconv"
)

func init() {
	RegisterMsg(CONNECT, &MsgConnect{})
	RegisterMsg(DISCONNECT, &MsgDisconnect{})
	//RegisterMsg(strconv.Itoa(60001), &MsgSync{}) // queueServer2loginServer

	RegisterMsg(strconv.Itoa(10010), &MsgHeartbeat{})
	RegisterMsg(strconv.Itoa(10011), &MsgPing{})
	RegisterMsg(strconv.Itoa(10013), &MsgLogin{})
}
