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

type Msger struct {
	Header byte
	Cmd int32
	Len int32
}

/*
|||||||||||||||||||||||||||||||||||||||||||||||
	公用通信协议,包含c2s,s2s
|||||||||||||||||||||||||||||||||||||||||||||||
 */
/*
	msg heartbeat c2queueserver queueserver2loginserver
	h:0x05, cmd:10010,len:0
 */
type MsgHeartbeat struct {
	Msger
}


/*
|||||||||||||||||||||||||||||||||||||||||||||||
	服务器间通信协议,客户端不可见
|||||||||||||||||||||||||||||||||||||||||||||||
 */
/*
	msg sync queueserver2loginserver
	h:0x05, cmd:11010,len:0
 */
type MsgSync struct {
	Msger
	Max int		// 服务器最大承载人数
	Cur int		// 当前服务器已连接人数
	Time int	// 每接入一个client所需时间s
}


/*
|||||||||||||||||||||||||||||||||||||||||||||||
	client与queueServer通信协议
|||||||||||||||||||||||||||||||||||||||||||||||
 */
/*
	msg ping client2queueserver
	h:0x05, cmd:10010,len:0
 */
type MsgPing struct {
	Msger
}

/*
	msg pong queueserver2client
	h:0x05, cmd:10010,len:0
 */
type MsgPong struct {
	Msger
	Time int	// 所需连入LoginServer的预估时间
	Queue int	// 当前服务器的排队人数
	InQueue	int	// 当前所在排队服务器的位置
}