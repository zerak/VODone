package msgs

import (
	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"

	"VODone/QueueServer/queue"
	"bitbucket.org/serverFramework/serverFramework/utils"
	"container/list"
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
	} else {
		ServerLogger.Info("not find client", client.GetID(), queue.ClientsMap)
	}
}
