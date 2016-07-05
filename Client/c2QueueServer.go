package main

import (
	"bufio"
	"sync"

	"VODone/Client/msgs"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/zhuangsirui/binpacker"
	"io"
	"net"
	"time"
)

var readerQueue *bufio.Reader
var writerQueue *bufio.Writer
var writeLockQueue sync.RWMutex
var MsgChanQueue chan *msgs.Message
var ExitChanQueue chan int
var PoolQueue sync.Pool

func startQueueServerLoop(conn net.Conn) {
	fmt.Printf("start Queue server goroutine\n")
	if _, err := Send2Queue(conn, []byte("  V1")); err != nil {
		fmt.Printf("send protocol err\n")
		panic(err)
	}

	wg.Wrap(func() {
		client2QueueServerLoop(conn)
	})
}

func client2QueueServerLoop(client net.Conn) {
	fmt.Printf("client2QueueServerLoop remoteAddr[%v] localAddr[%v]\n", client.RemoteAddr(), client.LocalAddr())
	var err error
	var header byte
	var cmd uint32
	var length uint32

	msgPumpStartedChan := make(chan bool)
	go clientMsgPumpQueue(client, msgPumpStartedChan)
	<-msgPumpStartedChan

	buf := make([]byte, ProtocolHeaderLen)
	for {
		_, err = io.ReadFull(readerQueue, buf)
		if err != nil {
			fmt.Printf("client2QueueServerLoop read head from remote[%v] err->%v buffed->%v\n", client.RemoteAddr(), err, readerQueue.Buffered())
			//ExitChanQueue <- 1
			break
		}

		// header
		header = buf[0]
		if header != 0x05 {
			err = fmt.Errorf("client2QueueServerLoop header[%s] err", header)
			//ExitChanQueue <- 1
			break
		}

		// cmd
		cmd = binary.BigEndian.Uint32(buf[1:5])

		// length
		length = binary.BigEndian.Uint32(buf[5:9])

		// data
		data := make([]byte, length)
		_, err = io.ReadFull(readerQueue, data)
		if err != nil {
			fmt.Printf("client2QueueServerLoop read data from client[%v] err->%v buffed->%v", client.RemoteAddr(), err, readerQueue.Buffered())
			//ExitChanQueue <- 1
			break
		}

		fmt.Printf("client2QueueServerLoop header[%v] cmd[%v] len[%d] data[%x]\n", header, cmd, length, data)

		// new msg
		//msg := Pool.Get().(*msgs.Message)
		//msg := &msgs.Message{ID:(int32)(cmd),Body:data,Conn:client}
		var msg msgs.Message
		msg.ID = int(cmd)
		msg.Body = data
		msg.Len = (int)(length)
		msg.Conn = client

		MsgChanQueue <- &msg
	}

	client.Close()

}

func clientMsgPumpQueue(client net.Conn, startedChan chan bool) {
	close(startedChan)

	hbTickerQueue := time.NewTicker(C2QueueServerHB)
	hbChanQueue := hbTickerQueue.C

	ppTickerQueue := time.NewTicker(C2QueueServerPP)
	ppChanQueue := ppTickerQueue.C
	for {
		select {
		case <-hbChanQueue:
			//var hb msgs.MsgHeartbeat
			//hb.Header = 0x05
			//hb.Cmd = 10010
			//hb.Len = 0
			buf := new(bytes.Buffer)
			packer := binpacker.NewPacker(buf, binary.BigEndian)
			packer.PushByte(0x05)
			packer.PushInt32(10010)
			packer.PushInt32(0)
			if err := packer.Error(); err != nil {
				fmt.Printf("clientMsgPumpQueue make msg err [%v]\n", err)
				ExitChanQueue <- 1
				break
			}

			fmt.Printf("clientMsgPumpQueue buf[%x] \n", buf.Bytes())

			if _, err := Send2Queue(client, buf.Bytes()); err != nil {
				fmt.Printf("clientMsgPumpQueue send packet err[%v] \n", err)
				ExitChanQueue <- 1
			}
		case <-ppChanQueue:
			//var hb msgs.MsgPing
			//hb.Header = 0x05
			//hb.Cmd = 10011
			//hb.Len = 0
			buf := new(bytes.Buffer)
			packer := binpacker.NewPacker(buf, binary.BigEndian)
			packer.PushByte(0x05)
			packer.PushInt32(10011)
			packer.PushInt32(0)
			if err := packer.Error(); err != nil {
				fmt.Printf("clientMsgPumpQueue make msgPing err [%v]\n", err)
				ExitChanQueue <- 1
			}

			fmt.Printf("clientMsgPumpQueue msgPing buf[%x] \n", buf.Bytes())

			if _, err := Send2Queue(client, buf.Bytes()); err != nil {
				fmt.Printf("clientMsgPumpQueue send packetPing err[%v] \n", err)
				ExitChanQueue <- 1
			}
		case msg, ok := <-MsgChanQueue:
			if ok {
				fmt.Printf("clientMsgPumpQueue msg[%v] body[%v]\n", msg.ID, msg.Body)
				if msg.ID == 10012 {
					buf := new(bytes.Buffer)
					packer := binpacker.NewPacker(buf,binary.BigEndian)
					packer.PushString(string(msg.Body[:]))
					unpacker := binpacker.NewUnpacker(buf,binary.BigEndian)
					var queue, inQueue, time int32
					if err := unpacker.FetchInt32(&queue).FetchInt32(&inQueue).FetchInt32(&time).Error(); err != nil{
						fmt.Printf("clientMsgPumpQueue unpacker msgPing err[%v]\n", err)
						ExitChanQueue <- 1
					}
					fmt.Printf("clientMsgPumpQueue msgPing queue[%v] inQueue[%v] time[%v]\n", queue, inQueue, time)
				}
			} else {
				fmt.Printf("clientMsgPumpQueue from MsgChan not ok\n")
				ExitChanQueue <- 1
			}
		case <-ExitChanQueue:
			goto exit
		}
	}

exit:
	client.Close()
	hbTickerQueue.Stop()
	ppTickerQueue.Stop()
	close(ExitChanQueue)
}

func Send2Queue(c net.Conn, data []byte) (int, error) {
	writeLockQueue.Lock()
	// todo

	// check write len(data) size buf
	n, err := writerQueue.Write(data)
	if err != nil {
		writeLockQueue.Unlock()
		return n, err
	}
	writerQueue.Flush()
	writeLockQueue.Unlock()

	return n, nil
}

func sendQueuePakcet(conn net.Conn) {
	// todo
	// 向QueueServer发送登录信息
	ticker := time.NewTicker(time.Second * 5)
	for _ = range ticker.C {
		buf := new(bytes.Buffer)
		packer := binpacker.NewPacker(buf, binary.BigEndian)
		packer.PushByte(0x05)
		packer.PushInt32(10013)
		accout := "account"
		passwd := "passwd"
		len := len(accout) + len(passwd)
		packer.PushInt32((int32)(len))
		packer.PushString(accout).PushString(passwd)
		if err := packer.Error(); err != nil {
			fmt.Printf("clientMsgPumpQueue make msg err [%v]\n", err)
			panic(err)
		}

		fmt.Printf("client send buf[%x] dataLen[%v]\n", buf.Bytes(), len)

		if _, err := Send2Queue(conn, buf.Bytes()); err != nil {
			fmt.Printf("clientMsgPumpQueue send c2sQueue packet err[%v] \n", err)
			panic(err)
		}

		ticker.Stop()
	}
}

func connect2QueueServer(addr string) net.Conn {
	fmt.Printf("connect2QueueServer [%v]\n", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	readerQueue = bufio.NewReaderSize(conn, defaultBufferSize)
	writerQueue = bufio.NewWriterSize(conn, defaultBufferSize)

	startQueueServerLoop(conn)

	return conn
}
