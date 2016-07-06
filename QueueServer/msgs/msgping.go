package msgs

import (
	"fmt"
	"bytes"
	"encoding/binary"

	"github.com/zhuangsirui/binpacker"

	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"

	"VODone/QueueServer/queue"
)

type MsgPing struct {
}

func (m *MsgPing) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("cid[%v] msg ping", client.GetID())

	buf := new(bytes.Buffer)
	packer := binpacker.NewPacker(buf, binary.BigEndian)
	packer.PushByte(0x05)
	packer.PushInt32(10012)
	len := 4 + 4 + 4	// queue inQueue time

	interVal := queue.GetInterVal()
	var que,inQueue,time int
	var flag byte
	que = queue.GetQueueClients()
	if que <= 0 {
		flag = '0'
	}
	inQueue = queue.GetQueuedIndex(client.GetID())
	time = (int)(interVal) * (inQueue + 1)

	packer.PushInt32((int32)(len))
	packer.PushInt32((int32)(que))
	packer.PushInt32((int32)(inQueue))
	packer.PushInt32((int32)(time))

	ServerLogger.Info("Queued clients flag[%v] len[%v] queue[%v] inQueue[%v] time[%v]", flag, queue.GetQueueClients(), que, inQueue, time)

	if _, err := p.Send(client, buf.Bytes()); err != nil {
		err = fmt.Errorf("failed to send response ->%s", err)
		client.Exit()
	}
}
