package msgs

import (
	"net"
	"sync"
	"time"

	"bitbucket.org/serverFramework/serverFramework/utils"
)

const (
	DefaultBufferSize = 16 * 1024
	ProtocolHeaderLen = 9 // header1 + cmd4 + length4

	C2LoginServerHB = 5 * time.Second
	C2QueueServerHB = 5 * time.Second
	C2QueueServerPP = 1 * time.Second
)

var WG utils.WaitGroupWrapper

type Message struct {
	ID        int
	Body      []byte
	Len       int
	Timestamp time.Time
	pool      sync.Pool

	net.Conn
}

/*
|||||||||||||||||||||||||||||||||||||||||||||||
		消息头
|||||||||||||||||||||||||||||||||||||||||||||||
*/
type Msger struct {
	Header byte
	Cmd    int32
	Len    int32
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
}

/*
|||||||||||||||||||||||||||||||||||||||||||||||
	服务器间通信协议,客户端不可见
|||||||||||||||||||||||||||||||||||||||||||||||
*/
/*
	msg sync queueserver2loginserver
	h:0x05, cmd:60000,len:0
	msg sync loginserver2queueserver
	h:0x05, cmd:60001,len:0
*/
type MsgSyncQ2L struct {
}
type MsgSyncL2Q struct {
	Max  int // 服务器最大承载人数
	Cur  int // 当前服务器已连接人数
	Time int // 每接入一个client所需时间s
}

/*
|||||||||||||||||||||||||||||||||||||||||||||||
	client与queueServer通信协议
|||||||||||||||||||||||||||||||||||||||||||||||
*/
/*
	msg ping client2queueserver
	h:0x05, cmd:10011,len:0
	msg pong queueserver2client
	h:0x05, cmd:10012,len:0
*/
type MsgPingC2Q struct {
}
type MsgPongQ2C struct {
	Queue   int // 当前服务器的排队人数
	InQueue int // 当前所在排队服务器的位置(返回排在前面的人数)
	Time    int // 所需连入LoginServer的预估时间
}

/*
|||||||||||||||||||||||||||||||||||||||||||||||
	client与LoginServer通信协议
|||||||||||||||||||||||||||||||||||||||||||||||
*/
/*
	msg logc2s client2loginserver
	h:0x05, cmd:10013,len:x
	msg logs2c loginserver2client
	h:0x05, cmd:10014,len:x
	flag/addr
	flag/uid/name
*/
type MsgLoginC2S struct {
	Account [512]byte // 帐号
	Passwd  [512]byte // 密码
}
type MsgLoginS2C struct {
	Flag byte // 登录成员与否标识 1成功,0失败
	Addr byte // 如果LoginServer负载高,返回QueueServer登录地址,"192.168.1.127:60060"
	UID  int64  // user id
	Name byte // 名字
}
