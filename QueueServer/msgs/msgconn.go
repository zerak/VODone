package msgs

import (
	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"

	"VODone/QueueServer/queue"
	"container/list"
)

type MsgConnect struct {
}

func (m *MsgConnect) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("MsgConnect cid[%v] msg connect", client.GetID())
	// add to queue list and map
	if _, ok := queue.ClientsMap[client.GetID()]; !ok {
		index := queue.QueuedClients.Len()
		qc := queue.QueueClient{ID: client.GetID(), Conn: client.GetConnector()}
		qc.SetIndex(index)

		queue.ClientsMap[client.GetID()] = &qc
		queue.QueuedClients.PushBack(&qc)
	}
}

type MsgDisconnect struct {
}

func (m *MsgDisconnect) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("MsgDisconnect cid[%v] msg disconnect", client.GetID())
	// todo
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
				element.Value.(*queue.QueueClient).SetIndex(index -1)
				ServerLogger.Info("MsgDisconnect update index cid[%v] addr[%v]", element.Value.(*queue.QueueClient).ID,element.Value.(*queue.QueueClient).Conn.LocalAddr().String())
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
