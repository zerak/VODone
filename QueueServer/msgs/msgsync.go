package msgs

import (
	"strconv"

	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"
)

type MsgSync struct {
}

func init() {
	RegisterMsg(strconv.Itoa(10010), &MsgSync{})
}

func (m *MsgSync) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("cid[%v] msg sync", client.GetID())
}
