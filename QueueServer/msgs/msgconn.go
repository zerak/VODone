package msgs

import (
	"net"
	"bytes"
	"container/list"
	"encoding/binary"

	"github.com/zhuangsirui/binpacker"

	"bitbucket.org/serverFramework/serverFramework/utils"
	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"

	"VODone/QueueServer/queue"
)

type MsgConnect struct {
}

func (m *MsgConnect) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("MsgConnect cid[%v] msg connect", client.GetID())
	// todo QueueServer UUID seesion
	// QueueServer里的UUID、应该从LoginServer提供的seesion队列里拿
	// 比如一次性生成几十G的数据放入redis数据库里

	// add to queue list and map
	if _, ok := queue.ClientsMap[client.GetID()]; !ok {
		index := queue.QueuedClients.Len()
		uuid, _ := utils.NewUUID()
		qc := queue.QueueClient{ID: client.GetID(), Conn: client.GetConnector(), Session: uuid}
		qc.SetIndex(index)

		queue.ClientsMap[client.GetID()] = &qc
		queue.QueuedClients.PushBack(&qc)
	}

	if err := notifyOtherClientQueueInfo(); err != nil {
		client.Exit()
	}
}

type MsgDisconnect struct {
}

func (m *MsgDisconnect) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("MsgDisconnect cid[%v] msg disconnect", client.GetID())
	// 1,remove from queue list and map
	// 2,update clients index
	if _, ok := queue.ClientsMap[client.GetID()]; ok {
		// remove
		update := false
		delete(queue.ClientsMap, client.GetID())

		var n *list.Element
		for element := queue.QueuedClients.Front(); element != nil; element = n {
			n = element.Next()
			if update {
				index := element.Value.(*queue.QueueClient).GetIndex()
				element.Value.(*queue.QueueClient).SetIndex(index - 1)
				ServerLogger.Info("MsgDisconnect update index cid[%v] addr[%v]", element.Value.(*queue.QueueClient).ID, element.Value.(*queue.QueueClient).Conn.LocalAddr().String())
			}

			if element.Value.(*queue.QueueClient).ID == client.GetID() {
				ServerLogger.Info("MsgDisconnect remove cid[%v] success", client.GetID())
				queue.QueuedClients.Remove(element)
				update = true
			}
		}

		if err := notifyOtherClientQueueInfo(); err != nil {
			client.Exit()
		}

		//// 如果有客户端断开连接,就检查是否通知其他排队的客户端
		//if queue.QueuedClients.Front() != nil && queue.AuthClients < queue.MaxClients {
		//	qc := queue.QueuedClients.Front().Value.(*queue.QueueClient)
		//
		//	buf := new(bytes.Buffer)
		//	packer := binpacker.NewPacker(buf, binary.BigEndian)
		//	packer.PushByte(0x05)
		//	packer.PushInt32(10012)
		//
		//	var flag byte
		//	uuid := qc.Session
		//	addr := "127.0.0.1:60060"
		//	len := 1 + len(uuid) + len(addr) // flag uuid addr
		//	flag = '1'
		//	packer.PushInt32((int32(len)))
		//	packer.PushByte(flag)
		//	packer.PushString(uuid)
		//	packer.PushString(addr)
		//
		//	if _, err := notice(qc.Conn, buf.Bytes()); err != nil {
		//	}
		//
		//	// TODO NOTE 向客户端发送重新登录消息,不管发送成功还是失败,都断开连接
		//	qc.Conn.Close()
		//}
	} else {
		ServerLogger.Info("not find client", client.GetID(), queue.ClientsMap)
	}
}

func notifyOtherClientQueueInfo() error {
	// 有客户端断连接,就通知其当前排队情况
	for element := queue.QueuedClients.Front(); element != nil; element = element.Next() {
		qc := element.Value.(*queue.QueueClient)

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
		inQueue = queue.GetQueuedIndex(qc.ID)
		time = (int)(interVal) * (inQueue + 1)

		packer.PushInt32((int32)(len))
		packer.PushByte(flag)
		packer.PushInt32((int32)(que))
		packer.PushInt32((int32)(inQueue))
		packer.PushInt32((int32)(time))

		if _, err := notice(qc.Conn, buf.Bytes()); err != nil {
			ServerLogger.Info("failed to send response ->%s", err)
			return err
		}
	}
	return nil
}

func notice(conn net.Conn, by []byte) (n int, err error) {
	if n, err := conn.Write(by); err != nil {
		return n, err
	}
	return n, nil
}
