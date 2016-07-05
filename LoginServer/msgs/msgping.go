package msgs

import (
	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"
)

type MsgPing struct {
}

func (m *MsgPing) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("cid[%v] msg ping", client.GetID())

}
