package msgs

import (
	"sync"
	"time"
	"net"
)

type Message struct {
	ID        int
	Body      []byte
	Timestamp time.Time
	pool      sync.Pool

	net.Conn
}

type MsgSync struct {

}

type MsgHeartbeat struct {
	Header byte
	Cmd int32
	Len int32
	Data []byte
}