package queue

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/zhuangsirui/binpacker"

	"VODone/Client/msgs"
)

var ReaderLoginQueue *bufio.Reader
var WriterLoginQueue *bufio.Writer
var writeLockLoginQueue sync.RWMutex
var MsgChanLoginQueue chan *msgs.Message
var ExitChanLoginQueue chan int
var PoolLoginQueue sync.Pool

func init() {
	MsgChanLoginQueue = make(chan *msgs.Message, 10000)
	PoolLoginQueue.New = func() interface{} {
		return &msgs.Message{
			Timestamp: time.Now(),
		}
	}
	ExitChanLoginQueue = make(chan int, 1)
}

func QueueStartLoginServerLoop(conn net.Conn) {
	fmt.Printf("LoginServer start goroutine\n")
	if _, err := QueueSend2Login(conn, []byte("  V1")); err != nil {
		fmt.Printf("send protocol err\n")
		panic(err)
	}

	msgs.WG.Wrap(func() {
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

	buf := make([]byte, msgs.ProtocolHeaderLen)
	for {
		_, err = io.ReadFull(ReaderLoginQueue, buf)
		if err != nil {
			fmt.Printf("client2LoginServerLoop read head from remote[%v] err->%v buffed->%v\n", client.RemoteAddr(), err, ReaderLoginQueue.Buffered())
			//ExitChanLoginQueue <- 1
			break
		}

		// header
		header = buf[0]
		if header != 0x05 {
			err = fmt.Errorf("client2LoginServerLoop header[%s] err", header)
			//ExitChanLoginQueue <- 1
			break
		}

		// cmd
		cmd = binary.BigEndian.Uint32(buf[1:5])

		// length
		length = binary.BigEndian.Uint32(buf[5:9])

		// data
		data := make([]byte, length)
		_, err = io.ReadFull(ReaderLoginQueue, data)
		if err != nil {
			fmt.Printf("client2LoginServerLoop read data from client[%v] err->%v buffed->%v", client.RemoteAddr(), err, ReaderLoginQueue.Buffered())
			//ExitChanLoginQueue <- 1
			break
		}

		//fmt.Printf("client2LoginServerLoop cid[%v] header[%v] cmd[%v] len[%d] data[%x]\n", client.LocalAddr().String(), header, cmd, length, data)

		// new msg
		//msg := Pool.Get().(*msgs.Message)
		//msg := &msgs.Message{ID:(int32)(cmd),Body:data,Conn:client}
		var msg msgs.Message
		msg.ID = int(cmd)
		msg.Body = data
		msg.Len = (int)(length)
		msg.Conn = client

		MsgChanLoginQueue <- &msg
	}

	client.Close()
	//ExitChanLoginQueue <- 1

	defer func() {
		fmt.Printf("client2LoginServerLoop cid[%v] exit\n", client.LocalAddr().String())
	}()
}

func clientMsgPumpLogin(client net.Conn, startedChan chan bool) {
	close(startedChan)

	hbTickerLogin := time.NewTicker(msgs.C2LoginServerHB)
	hbChanLogin := hbTickerLogin.C
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
				ExitChanLoginQueue <- 1
			}

			//fmt.Printf("clientMsgPumpLogin heartbeat cid[%v] buf[%x] \n", client.LocalAddr().String(), buf.Bytes())

			if n, err := QueueSend2Login(client, buf.Bytes()); err != nil || n != 9 {
				fmt.Printf("clientMsgPumpLogin send heartbeat packet err[%v] \n", err)
				ExitChanLoginQueue <- 1
			} else {
				//fmt.Printf("msg hb len[%v]\n", n)
			}
		case msg, ok := <-MsgChanLoginQueue:
			if ok {
				//fmt.Printf("clientMsgPumpLogin cid[%v] msgChan msg[%v] body[%v]\n", client.LocalAddr().String(), msg.ID, msg.Body)
				if msg.ID == 10014 {
					buf := new(bytes.Buffer)
					packer := binpacker.NewPacker(buf, binary.BigEndian)
					packer.PushString(string(msg.Body[:]))
					unpacker := binpacker.NewUnpacker(buf, binary.BigEndian)

					var flag byte
					if err := unpacker.FetchByte(&flag).Error(); err != nil {
						fmt.Printf("clientMsgPumpLogin unpacker err[%v]\n", err)
						ExitChanLoginQueue <- 1
					}

					//fmt.Printf("clientMsgPumpLogin cid[%v] flag[%v]\n", client.LocalAddr().String(), flag)
					if flag == 48 {
						//todo
						// login server return err, and connect to queue server
						var addr string
						len := uint64(msg.Len - 1)
						if err := unpacker.FetchString(len, &addr).Error(); err != nil {
							fmt.Printf("clientMsgPumpLogin login failed and get queue server addr err\n")
						}
						fmt.Printf("clientMsgPumpLogin login failed and redirect to queue server[%v]\n", addr)
						Connect2QueueServer(addr)
						ExitChanLoginQueue <- 1
					} else {
						var uid int64
						var name string
						len := uint64(msg.Len - 1 - 8)
						if err := unpacker.FetchInt64(&uid).FetchString(len, &name).Error(); err != nil {
							fmt.Printf("clientMsgPumpLogin login success but unpack failed [%v]\n", err)
							ExitChanLoginQueue <- 1
						}
						fmt.Printf("clientMsgPumpLogin login success uid[%v] name[%v]\n", uid, name)
					}
				}
			} else {
				fmt.Printf("clientMsgPumpLogin from MsgChan not ok\n")
				ExitChanLoginQueue <- 1
			}
		case <-ExitChanLoginQueue:
			fmt.Printf("clientMsgPumpLogin exitChan recv EXIT\n")
			goto exit
		}
	}

exit:
	client.Close()
	hbTickerLogin.Stop()
	close(ExitChanLoginQueue)

	defer func() {
		fmt.Printf("clientMsgPumpLogin exit\n")
	}()
}

func QueueSend2Login(c net.Conn, data []byte) (int, error) {
	writeLockLoginQueue.Lock()
	// todo

	// check write len(data) size buf
	n, err := WriterLoginQueue.Write(data)
	if err != nil {
		writeLockLoginQueue.Unlock()
		return n, err
	}
	WriterLoginQueue.Flush()
	writeLockLoginQueue.Unlock()

	return n, nil
}

func SendLoginPakcetWithKey(conn net.Conn, uuid string) {
	// todo
	// 向LoginServer发送登录信息
	buf := new(bytes.Buffer)
	packer := binpacker.NewPacker(buf, binary.BigEndian)
	packer.PushByte(0x05)
	packer.PushInt32(10013)
	var flag byte
	flag = '1'
	accout := "account"
	passwd := "passwd"
	key := uuid
	len := 1 + len(accout) + len(passwd) + len(uuid)
	packer.PushInt32((int32)(len))
	packer.PushByte(flag)
	packer.PushString(key)
	packer.PushString(accout).PushString(passwd)
	if err := packer.Error(); err != nil {
		fmt.Printf("make msg err [%v]\n", err)
		panic(err)
	}

	fmt.Printf("client reconnect to loginserver c2slogin packet buf[%x] dataLen[%v]\n", buf.Bytes(), len)

	if _, err := QueueSend2Login(conn, buf.Bytes()); err != nil {
		fmt.Printf("send c2slogin packet err[%v] \n", err)
		panic(err)
	}
}

func QueueConnect2LoginServer(addr string) net.Conn {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	ReaderLoginQueue = bufio.NewReaderSize(conn, msgs.DefaultBufferSize)
	WriterLoginQueue = bufio.NewWriterSize(conn, msgs.DefaultBufferSize)

	QueueStartLoginServerLoop(conn)
	return conn
}
