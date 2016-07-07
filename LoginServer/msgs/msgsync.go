package msgs

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/zhuangsirui/binpacker"

	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"
)

type MsgSync struct {
}

func (m *MsgSync) ProcessMsg(p Protocol, client Client, msg *Message) {
	//ServerLogger.Info("cid[%v] msg sync", client.GetID())

	QueueServerIdentify = client.String()
	//ServerLogger.Warn("QueueServer[%v]", client.String())

	buf := new(bytes.Buffer)
	packer := binpacker.NewPacker(buf, binary.BigEndian)
	packer.PushByte(0x05)
	packer.PushInt32(60001)
	Max := ServerApp.GetMaxClients()
	Cur := ServerApp.GetAuthClients()
	Time := 10
	//len := len(int32(Max)) + len(int32(Cur)) + len(int32(Time))
	len := 4 + 4 + 4
	packer.PushInt32((int32)(len))

	packer.PushInt32(int32(Max))
	packer.PushInt32(int32(Cur))
	packer.PushInt32(int32(Time))

	if err := packer.Error(); err != nil {
		ServerLogger.Info("make msg err [%v]", err)
		client.Exit()
	}

	//ServerLogger.Info("msgSync login2queue buf[%x] len[%v] max[%v] auth[%v]", buf.Bytes(), len, Max, Cur)
	if _, err := p.Send(client, buf.Bytes()); err != nil {
		err = fmt.Errorf("failed to send response ->%s", err)
		client.Exit()
	}
}
