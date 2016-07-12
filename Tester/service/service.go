package service

import (
	"bufio"
	"net"
	"sync"
)

type Service interface {
	GetReader() *bufio.Reader
	GetWriter() *bufio.Writer

	GetConn() net.Conn
	GetLock() sync.RWMutex
}

type ServiceLogin interface {
	Service
	Connect2LoginServer(addr string) error
	Connect2QueueServer(addr string) error
}

type ServiceQueue interface {
	Service
	Connect2LoginServer(addr string) error
	Connect2QueueServer(addr string) error
}
