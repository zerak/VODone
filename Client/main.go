package main

import (
	"net"
	"fmt"
	"io"
	"bufio"
	"time"
	"sync"
	"encoding/binary"

	"bitbucket.org/serverFramework/serverFramework/utils"
	"VODone/Client/msgs"
	"bytes"
)

const (
	defaultBufferSize = 16 * 1024
	ProtocolHeaderLen = 9 // header1 + cmd4 + length4

	C2LoginServerHB = 5 * time.Second
	C2QueueServerHB = 5 * time.Second
)

var writeLock sync.RWMutex
var reader *bufio.Reader
var writer *bufio.Writer
var MsgChan chan *msgs.Message
var ExitChan chan byte
var Pool sync.Pool

func init(){
	Pool.New = func() interface{} {
		return &msgs.Message{
			Timestamp: time.Now(),
		}
	}
}

func main() {
	// first connect to login server
	conn, err := net.Dial("tcp", "127.0.0.1:60060")
	if err != nil {
		panic(err)
	}

	reader = bufio.NewReaderSize(conn, defaultBufferSize)
	writer = bufio.NewWriterSize(conn,defaultBufferSize)

	if _, err := Send(conn,[]byte("  V1")); err!= nil{
		fmt.Printf("send protocle err\n")
		panic(err)
	}

	fmt.Printf("start goroutine\n")
	var wg utils.WaitGroupWrapper
	wg.Wrap(func() {
		clientLoop(conn)
	})

	wg.Wait()

	fmt.Printf("client exit\n")
}

func clientLoop(client net.Conn) {
	fmt.Printf("remoteAddr[%v] localAddr[%v]\n",client.RemoteAddr(),client.LocalAddr())
	var err error
	var header byte
	var cmd uint32
	var length uint32

	msgPumpStartedChan := make(chan bool)
	go clientMsgPump(client,msgPumpStartedChan)
	<-msgPumpStartedChan

	buf := make([]byte, ProtocolHeaderLen)
	for {
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			fmt.Printf("clientLoop read head from remote[%v] err->%v buffed->%v\n", client.RemoteAddr(), err, reader.Buffered())
			break
		}

		// header
		header = buf[0]
		if header != 0x05 {
			err = fmt.Errorf("clientLoop header[%s] err", header)
			break
		}

		// cmd
		cmd = binary.BigEndian.Uint32(buf[1:5])

		// length
		length = binary.BigEndian.Uint32(buf[5:9])

		// data
		data := make([]byte, length)
		_, err = io.ReadFull(reader, data)
		if err != nil {
			fmt.Printf("clientLoop read data from client[%v] err->%v buffed->%v", client.RemoteAddr(), err, reader.Buffered())
			break
		}

		fmt.Printf("clientLoop header[%v] cmd[%v] len[%d] data[%v]", header, cmd, length, string(data))

		// new msg
		msg := Pool.Get().(*msgs.Message)
		//msg := NewMessage(int(cmd), data, client)
		msg.ID = int(cmd)
		msg.Body = data
		msg.Conn = client

		MsgChan <- msg
	}

	client.Close()
	ExitChan <- 1
}

func clientMsgPump(client net.Conn, startedChan chan bool) {
	close(startedChan)

	hbTickerLogin := time.NewTicker(C2LoginServerHB)
	hbChanLogin := hbTickerLogin.C

	hbTickerQueue := time.NewTicker(C2QueueServerHB)
	hbChanQueue := hbTickerQueue.C
	for {
		select {
		case <- hbChanLogin:
			hb := new(msgs.MsgHeartbeat)
			hb.Header = 0x05
			hb.Cmd = 10010
			hb.Len = 0
			strHb := fmt.Sprintf("%c%4d%4d",hb.Header, hb.Cmd, hb.Len)
			buf := new(bytes.Buffer)
			if err := binary.Write(buf, binary.BigEndian, hb); err!=nil{
			}
			fmt.Printf("buf[%x] \n",buf.Bytes())

			if _, err := Send(client,buf.Bytes()) ; err!=nil{
				goto exit
			}
			fmt.Printf("c2s heartbeat str[%v] by[%v]\n",strHb, []byte(strHb))
			//if _, err := Send(client,[]byte("C2LoginServerHB")) ; err!=nil{
			//	goto exit
			//}
		case <- hbChanQueue:
			//if _, err := Send(client,[]byte("C2QueueServerHB")) ; err!=nil{
			//	goto exit
			//}
		case msg, ok := <-MsgChan:
			if ok {
				fmt.Printf("msg[%v]\n",msg)
			}else {
				goto exit
			}
		case <- ExitChan:
			goto exit
		}
	}

exit:
	hbTickerLogin.Stop()
	hbTickerQueue.Stop()
	close(ExitChan)
}

func Send(c net.Conn, data []byte) (int, error) {
	writeLock.Lock()
	// todo

	// check write len(data) size buf
	n, err := writer.Write(data)
	if err != nil {
		writeLock.Unlock()
		return n, err
	}
	writer.Flush()
	writeLock.Unlock()

	return n, nil
}