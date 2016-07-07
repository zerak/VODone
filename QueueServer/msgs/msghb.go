package msgs

import (
	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"
)

type MsgHeartbeat struct {
}

//
//func init() {
//	//RegisterMsg(strconv.Itoa(10010), &MsgHeartbeat{})
//}

func (m *MsgHeartbeat) ProcessMsg(p Protocol, client Client, msg *Message) {
	//ServerLogger.Info("cid[%v] msg heartbeat", client.GetID())

}
