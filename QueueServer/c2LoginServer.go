package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"VODone/Client/msgs"
	"github.com/zhuangsirui/binpacker"
)

const (
	defaultBufferSize = 16 * 1024
	ProtocolHeaderLen = 9 // header1 + cmd4 + length4

	QueueServer2LoginServerHB = 5 * time.Second
	QueueServer2LoginServerPP = 1 * time.Second
)

var addr string

var readerLogin *bufio.Reader
var writerLogin *bufio.Writer
var writeLockLogin sync.RWMutex
var MsgChanLogin chan *msgs.Message
var ExitChanLogin chan int
var PoolLogin sync.Pool

func init() {
	MsgChanLogin = make(chan *msgs.Message, 10000)
	PoolLogin.New = func() interface{} {
		return &msgs.Message{
			Timestamp: time.Now(),
		}
	}
	ExitChanLogin = make(chan int, 1)
}
func startLoginServerLoop(conn net.Conn) {
	fmt.Printf("LoginServer start goroutine\n")
	if _, err := Send2Login(conn, []byte("  V1")); err != nil {
		fmt.Printf("send protocol err\n")
		panic(err)
	}

	wg.Wrap(func() {
		client2LoginServerLoop(conn)
	})
}

func client2LoginServerLoop(client net.Conn) {
	fmt.Printf("client2LoginServerLoop remoteAddr[%v] localAddr[%v]\n", client.RemoteAddr(), client.LocalAddr())
	var err error
	var header byte
	var cmd uint32
	var length uint32

	msgPumpStartedChan := make(chan bool)
	go clientMsgPumpLogin(client, msgPumpStartedChan)
	<-msgPumpStartedChan

	buf := make([]byte, ProtocolHeaderLen)
	for {
		_, err = io.ReadFull(readerLogin, buf)
		if err != nil {
			fmt.Printf("client2LoginServerLoop read head from remote[%v] err->%v buffed->%v\n", client.RemoteAddr(), err, readerLogin.Buffered())
			//ExitChanLogin <- 1
			break
		}

		// header
		header = buf[0]
		if header != 0x05 {
			err = fmt.Errorf("client2LoginServerLoop header[%s] err", header)
			//ExitChanLogin <- 1
			break
		}

		// cmd
		cmd = binary.BigEndian.Uint32(buf[1:5])

		// length
		length = binary.BigEndian.Uint32(buf[5:9])

		// data
		data := make([]byte, length)
		_, err = io.ReadFull(readerLogin, data)
		if err != nil {
			fmt.Printf("client2LoginServerLoop read data from client[%v] err->%v buffed->%v", client.RemoteAddr(), err, readerLogin.Buffered())
			//ExitChanLogin <- 1
			break
		}

		fmt.Printf("client2LoginServerLoop header[%v] cmd[%v] len[%d] data[%x]\n", header, cmd, length, data)

		// new msg
		//msg := Pool.Get().(*msgs.Message)
		//msg := &msgs.Message{ID:(int32)(cmd),Body:data,Conn:client}
		var msg msgs.Message
		msg.ID = int(cmd)
		msg.Body = data
		msg.Len = (int)(length)
		msg.Conn = client

		MsgChanLogin <- &msg
	}

	client.Close()
	//ExitChanLogin <- 1

	defer func() {
		fmt.Printf("client2LoginServerLoop exit\n")
	}()
}

func clientMsgPumpLogin(client net.Conn, startedChan chan bool) {
	close(startedChan)

	hbTickerLogin := time.NewTicker(QueueServer2LoginServerHB)
	hbChanLogin := hbTickerLogin.C

	syncTickerLogin := time.NewTicker(QueueServer2LoginServerPP)
	syncChanLogin := syncTickerLogin.C
	for {
		select {
		case <-hbChanLogin:
			buf := new(bytes.Buffer)
			packer := binpacker.NewPacker(buf, binary.BigEndian)
			packer.PushByte(0x05)
			packer.PushInt32(10010)
			packer.PushInt32(0)
			if err := packer.Error(); err != nil {
				fmt.Printf("clientMsgPumpLogin make msg err [%v]\n", err)
				ExitChanLogin <- 1
			}

			fmt.Printf("clientMsgPumpLogin heartbeat buf[%x] \n", buf.Bytes())

			if _, err := Send2Login(client, buf.Bytes()); err != nil {
				fmt.Printf("clientMsgPumpLogin send heartbeat packet err[%v] \n", err)
				ExitChanLogin <- 1
			}
		case <-syncChanLogin:
			//var hb msgs.MsgSync
			//hb.Header = 0x05
			//hb.Cmd = 60000
			//hb.Len = 0
			buf := new(bytes.Buffer)
			packer := binpacker.NewPacker(buf, binary.BigEndian)
			packer.PushByte(0x05)
			packer.PushInt32(60000)
			packer.PushInt32(0)
			if err := packer.Error(); err != nil {
				fmt.Printf("clientMsgPumpQueue make msgSync err [%v]\n", err)
				ExitChanLogin <- 1
			}

			fmt.Printf("clientMsgPumpQueue msgSync buf[%x] \n", buf.Bytes())

			if _, err := Send2Login(client, buf.Bytes()); err != nil {
				fmt.Printf("clientMsgPumpQueue send msgSync err[%v] \n", err)
				ExitChanLogin <- 1
			}
		case msg, ok := <-MsgChanLogin:
			if ok {
				fmt.Printf("clientMsgPumpLogin msgChan msg[%v] body[%v]\n", msg.ID, msg.Body)
				if msg.ID == 60001 {
					buf := new(bytes.Buffer)
					packer := binpacker.NewPacker(buf, binary.BigEndian)
					packer.PushString(string(msg.Body[:]))
					unpacker := binpacker.NewUnpacker(buf, binary.BigEndian)
					var max, cur, time int32
					unpacker.FetchInt32(&max).FetchInt32(&cur).FetchInt32(&time)
					if err := unpacker.Error(); err != nil {
						ExitChanLogin <- 1
					}
					SetMaxClients(max)
					SetCurClients(cur)
					SetInterVal(time)
					fmt.Printf("clientMsgPumpLogin msgsync login2queue max[%v] cur[%v] time[%v]\n", max, cur, time)
				}

			} else {
				fmt.Printf("clientMsgPumpLogin from MsgChan not ok\n")
				ExitChanLogin <- 1
			}
		case <-ExitChanLogin:
			fmt.Printf("clientMsgPumpLogin exitChan recv EXIT\n")
			goto exit
		}
	}

exit:
	client.Close()
	hbTickerLogin.Stop()
	close(ExitChanLogin)

	defer func() {
		fmt.Printf("clientMsgPumpLogin exit\n")
	}()
}

func Send2Login(c net.Conn, data []byte) (int, error) {
	writeLockLogin.Lock()
	// todo

	// check write len(data) size buf
	n, err := writerLogin.Write(data)
	if err != nil {
		writeLockLogin.Unlock()
		return n, err
	}
	writerLogin.Flush()
	writeLockLogin.Unlock()

	return n, nil
}

func connect2LoginServer(addr string) net.Conn {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	readerLogin = bufio.NewReaderSize(conn, defaultBufferSize)
	writerLogin = bufio.NewWriterSize(conn, defaultBufferSize)

	startLoginServerLoop(conn)
	return conn
}
