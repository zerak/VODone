package msgs

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/zhuangsirui/binpacker"

	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"

	"VODone/QueueServer/queue"
)

type MsgPing struct {
}

func (m *MsgPing) ProcessMsg(p Protocol, client Client, msg *Message) {
	//ServerLogger.Info("MsgPing cid[%v] msg ping", client.GetID())

	buf := new(bytes.Buffer)
	packer := binpacker.NewPacker(buf, binary.BigEndian)
	packer.PushByte(0x05)
	packer.PushInt32(10012)
	len := 1 + 4 + 4 + 4 // flag + queue inQueue time

	interVal := queue.GetInterVal()
	var que, inQueue, time int
	var flag byte
	que = queue.GetQueueClients()
	if que <= 0 {
		flag = '0'
	}
	inQueue = queue.GetQueuedIndex(client.GetID())
	time = (int)(interVal) * (inQueue + 1)

	packer.PushInt32((int32)(len))
	packer.PushByte(flag)
	packer.PushInt32((int32)(que))
	packer.PushInt32((int32)(inQueue))
	packer.PushInt32((int32)(time))

	if _, err := p.Send(client, buf.Bytes()); err != nil {
		err = fmt.Errorf("failed to send response ->%s", err)
		client.Exit()
	}

	// 通知开启Timer
	// 开启Timer,用于通知排队的客户端何时登录
	queue.TimerSig <- 1

	//ServerLogger.Info("MsgPing Queued clients flag[%v] len[%v] queue[%v] inQueue[%v] time[%v]", flag, queue.GetQueueClients(), que, inQueue, time)
}
