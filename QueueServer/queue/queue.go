package queue

import (
	"container/list"

	. "bitbucket.org/serverFramework/serverFramework/client"
)

var QueuedClients *list.List
var ClientsMap map[int64]*Client

func init() {
	QueuedClients = list.New()
	ClientsMap = make(map[int64]*Client)
}

func GetQueueClients() int {
	return QueuedClients.Len()
}

func GetQueuedIndex(c *Client) int {
	return 10
}