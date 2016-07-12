package msgs

import (
	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"
)

type MsgConnect struct {
}

func (m *MsgConnect) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("MsgConnect cid[%v] msg connect auth[%v]", client.GetID(), client.GetAuth())
}

type MsgDisconnect struct {
}

func (m *MsgDisconnect) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("MsgDisconnect cid[%v][%v][%v] auth[%v] msg disconnect", client.GetID(), client.GetIdentify(), client.String(), client.GetAuth())

	if client.String() != QueueServerIdentify && client.GetAuth() {
		ServerLogger.Warn("MsgDisconnect cid[%v][%v] sub auth", client.String(), client.GetIdentify())
		client.SetAuth(false)
		ServerApp.AuthClient(false)
	}
}
