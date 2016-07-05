package msgs

import (
	"bytes"
	"encoding/binary"

	"github.com/zhuangsirui/binpacker"

	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"

	"VODone/QueueServer/queue"
	"fmt"
)

type MsgPing struct {
}

func (m *MsgPing) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("cid[%v] msg ping", client.GetID())

	buf := new(bytes.Buffer)
	packer := binpacker.NewPacker(buf, binary.BigEndian)
	packer.PushByte(0x05)
	packer.PushInt32(10012)
	len := 4 + 4 + 4
	var q,inQueue,time int
	q = queue.GetQueueClients()
	inQueue = queue.GetQueuedIndex(&client)
	time = 10

	packer.PushInt32((int32)(len))
	packer.PushInt32((int32)(q))
	packer.PushInt32((int32)(inQueue))
	packer.PushInt32((int32)(time))

	if _, ok := queue.ClientsMap[client.GetID()]; !ok {
		queue.ClientsMap[client.GetID()] = &client
		queue.QueuedClients.PushBack(&client)
	}
	ServerLogger.Info("Queued clients len[%v] queue[%v] inQueue[%v] time[%v]", queue.QueuedClients.Len(), q, inQueue, time)

	if _, err := p.Send(client, buf.Bytes()); err != nil {
		err = fmt.Errorf("failed to send response ->%s", err)
		client.Exit()
	}
}
