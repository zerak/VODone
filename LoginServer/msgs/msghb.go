package msgs

import (
	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"
)

type MsgHeartbeat struct {
}

func (m *MsgHeartbeat) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("cid[%v] msg heartbeat", client.GetID())

}
