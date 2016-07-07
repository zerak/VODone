package queue

import (
	"bufio"
	"net"
	"strconv"
	"sync"
	"time"

	"bitbucket.org/serverFramework/serverFramework/core"
)

type QueueClient struct {
	ClientID string
	HostName string

	net.Conn

	ID int64

	// reading/writing interfaces
	Reader *bufio.Reader
	Writer *bufio.Writer

	HeartbeatInterval time.Duration
	MsgTimeout        time.Duration

	writeLock sync.RWMutex
	metaLock  sync.RWMutex
	index     int    // the index of the queue list, begin of 0
	Session   string // the queued client ticket
}

func (qc *QueueClient) String() string {
	return qc.RemoteAddr().String()
}

func (qc *QueueClient) Exit() {

}

func (qc *QueueClient) WLock() {

}

func (qc *QueueClient) WUnlock() {

}

func (qc *QueueClient) Write(data []byte) (int, error) {
	return qc.Writer.Write(data)
}

func (qc *QueueClient) Flush() {

}

func (qc *QueueClient) GetID() int64 {
	return qc.ID
}

func (qc *QueueClient) GetHBInterval() time.Duration {
	return qc.HeartbeatInterval
}

func (qc *QueueClient) GetConnector() net.Conn {
	return qc.Conn
}

func (qc *QueueClient) SetIndex(index int) {
	qc.index = index
}

func (qc *QueueClient) GetIndex() int {
	return qc.index
}

func (qc *QueueClient) Notify(uid int64, tpe string) {
	id, _ := strconv.Atoi(qc.ClientID)
	if uid != (int64)(id) {
		return
	}

	if tpe == core.CONNECT {
		// add to queue list and map
		if _, ok := ClientsMap[uid]; !ok {
			index := QueuedClients.Len()
			qc.SetIndex(index)
			ClientsMap[qc.ID] = qc
			QueuedClients.PushBack(qc)
		}
	} else if tpe == core.DISCONNECT {
		// remove from queue list and map
		if _, ok := ClientsMap[uid]; !ok {
			delete(ClientsMap, uid)
			for element := QueuedClients.Front(); element != nil; element = element.Next() {
				if element.Value == uid {
					QueuedClients.Remove(element)
					break
				}
			}
		}
	}
}
