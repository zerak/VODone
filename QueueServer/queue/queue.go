package queue

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
)

var QueuedClients *list.List
var ClientsMap map[int64]*QueueClient
var queueLock sync.Mutex

var MaxClients int32
var CurClients int32
var InterVal int32      // 时间间隔,loginServer返回的可登录时间间隔
var Timer int           // 排在最前的客户端所需连入服务器时间
var TimerReset chan int // 用于更新Timer

func init() {
	QueuedClients = list.New()
	ClientsMap = make(map[int64]*QueueClient)
	Timer = 10	// 初始值10秒、QueueServer启动10秒后,检查当前队列发送重登录消息
	TimerReset = make(chan int, 1)
}

func GetQueueClients() int {
	queueLock.Lock()
	defer queueLock.Unlock()
	return QueuedClients.Len()
}

func GetQueuedIndex(uid int64) int {
	fmt.Printf("queue GetQueuedIndex[%v]\n", uid)
	queueLock.Lock()
	defer queueLock.Unlock()
	if v, ok := ClientsMap[uid]; ok {
		fmt.Printf("queue GetQueuedIndex[%v] v[%v]\n", uid, v)
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

func SetCurClients(n int32) {
	atomic.StoreInt32(&CurClients, n)
}

func GetCurClients() int32 {
	return atomic.LoadInt32(&CurClients)
}

func SetInterVal(n int32) {
	atomic.StoreInt32(&InterVal, n)
}

func GetInterVal() int32 {
	return atomic.LoadInt32(&InterVal)
}
