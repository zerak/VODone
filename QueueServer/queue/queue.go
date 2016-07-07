package queue

import (
	"container/list"
	"sync"
	"sync/atomic"
)

var QueuedClients *list.List
var ClientsMap map[int64]*QueueClient
var queueLock sync.Mutex

var MaxClients int32
var AuthClients int32   // LoginServer已经成功登录的人数
var InterVal int32      // 时间间隔,loginServer返回的可登录时间间隔
var Timer int           // 排在最前的客户端所需连入服务器时间
var TimerSig chan int   // 用于通知开启Timer
var NotifyChan chan int // 用于向客户端发送通知

func init() {
	QueuedClients = list.New()
	ClientsMap = make(map[int64]*QueueClient)
	Timer = 10 // 初始值10秒、QueueServer启动10秒后,检查当前队列发送重登录消息
	TimerSig = make(chan int, 1)
	NotifyChan = make(chan int, 1)
}

func GetQueueClients() int {
	queueLock.Lock()
	defer queueLock.Unlock()
	return QueuedClients.Len()
}

func GetQueuedIndex(uid int64) int {
	queueLock.Lock()
	defer queueLock.Unlock()
	if v, ok := ClientsMap[uid]; ok {
		return v.index
	}
	return -1
}

func SetMaxClients(n int32) {
	atomic.StoreInt32(&MaxClients, n)
}

func GetMaxClients() int32 {
	return atomic.LoadInt32(&MaxClients)
}

func SetAuthClients(n int32) {
	atomic.StoreInt32(&AuthClients, n)
}

func GetCurClients() int32 {
	return atomic.LoadInt32(&AuthClients)
}

func SetInterVal(n int32) {
	atomic.StoreInt32(&InterVal, n)
}

func GetInterVal() int32 {
	return atomic.LoadInt32(&InterVal)
}
