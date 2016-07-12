package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"bitbucket.org/serverFramework/serverFramework/utils"
	"github.com/zhuangsirui/binpacker"

	"VODone/Tester/msgs"
	"VODone/Tester/service"
)

type QueueServer struct {
	login service.ServiceLogin

	reader    *bufio.Reader
	writer    *bufio.Writer
	writeLock sync.RWMutex

	MsgChan  chan *msgs.Message
	ExitChan chan int
	Conn     net.Conn
	Pool     sync.Pool
	Addr     string

	wg utils.WaitGroupWrapper
}

func (qs *QueueServer) SetLoginServerPtr(login service.ServiceLogin) {
	qs.login = login
}

func (qs *QueueServer) GetReader() *bufio.Reader {
	return qs.reader
}

func (qs *QueueServer) GetWriter() *bufio.Writer {
	return qs.writer
}

func (qs *QueueServer) GetConn() net.Conn {
	return qs.Conn
}

func (qs *QueueServer) GetLock() sync.RWMutex {
	return qs.writeLock
}

func (qs *QueueServer) Start() error {
	if _, err := msgs.Send(qs, []byte("  V1")); err != nil {
		return fmt.Errorf("send protocol [  v1]err [%v]\n", err)
	}

	return nil
}

func (qs *QueueServer) Run() {
	qs.wg.Wrap(func() {
		qs.serverLoop()
	})
}

func (qs *QueueServer) Stop() {
	close(qs.ExitChan)
	qs.Conn.Close()
	qs.wg.Wait()
}

func (qs *QueueServer) serverLoop() {
	var err error
	var header byte
	var cmd uint32
	var length uint32

	msgPumpStartedChan := make(chan bool)
	go qs.msgPump(msgPumpStartedChan)
	<-msgPumpStartedChan

	fmt.Printf("QueueServerLoop remoteAddr[%v] localAddr[%v]\n", qs.Conn.RemoteAddr(), qs.Conn.LocalAddr())

	buf := make([]byte, msgs.ProtocolHeaderLen)
	for {
		_, err = io.ReadFull(qs.reader, buf)
		if err != nil {
			fmt.Printf("QueueServerLoop read head from remote[%v] err->%v buffed->%v\n", qs.Conn.RemoteAddr(), err, qs.reader.Buffered())
			//ExitChan <- 1
			break
		}

		// header
		header = buf[0]
		if header != 0x05 {
			err = fmt.Errorf("QueueServerLoop header[%s] err", header)
			//ExitChan <- 1
			break
		}

		// cmd
		cmd = binary.BigEndian.Uint32(buf[1:5])

		// length
		length = binary.BigEndian.Uint32(buf[5:9])

		// data
		data := make([]byte, length)
		_, err = io.ReadFull(qs.reader, data)
		if err != nil {
			fmt.Printf("QueueServerLoop read data from client[%v] err->%v buffed->%v", qs.Conn.RemoteAddr(), err, qs.reader.Buffered())
			//ExitChan <- 1
			break
		}

		//fmt.Printf("QueueServerLoop cid[%v] header[%v] cmd[%v] len[%d] data[%x]\n", client.LocalAddr().String(), header, cmd, length, data)

		// new msg
		//msg := Pool.Get().(*msgs.Message)
		//msg := &msgs.Message{ID:(int32)(cmd),Body:data,Conn:client}
		var msg msgs.Message
		msg.ID = int(cmd)
		msg.Body = data
		msg.Len = (int)(length)
		msg.Conn = qs.Conn

		qs.MsgChan <- &msg
	}

	defer func() {
		qs.Conn.Close()
		//qs.ExitChan <- 1
		fmt.Printf("QueueServerLoop cid[%v] exit\n", qs.Conn.LocalAddr().String())
	}()
}

func (qs *QueueServer) msgPump(startedChan chan bool) {
	close(startedChan)

	hbTicker := time.NewTicker(msgs.C2QueueServerHB)
	hbChan := hbTicker.C

	ppTicker := time.NewTicker(msgs.C2QueueServerPP)
	ppChan := ppTicker.C
	for {
		select {
		case <-hbChan:
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
				fmt.Printf("QueueServerMsgPump make msg err [%v]\n", err)
				qs.ExitChan <- 1
				break
			}

			//fmt.Printf("QueueServerMsgPump buf[%x] \n", buf.Bytes())

			if _, err := msgs.Send(qs, buf.Bytes()); err != nil {
				fmt.Printf("QueueServerMsgPump send packet err[%v] \n", err)
				qs.ExitChan <- 1
			}
		case <-ppChan:
		////var hb msgs.MsgPing
		////hb.Header = 0x05
		////hb.Cmd = 10011
		////hb.Len = 0
		//buf := new(bytes.Buffer)
		//packer := binpacker.NewPacker(buf, binary.BigEndian)
		//packer.PushByte(0x05)
		//packer.PushInt32(10011)
		//packer.PushInt32(0)
		//if err := packer.Error(); err != nil {
		//	fmt.Printf("QueueServerMsgPump make msgPing err [%v]\n", err)
		//	ExitChanQueue <- 1
		//}
		//
		////fmt.Printf("QueueServerMsgPump msgPing buf[%x] \n", buf.Bytes())
		//
		//if _, err := Send2Queue(client, buf.Bytes()); err != nil {
		//	fmt.Printf("QueueServerMsgPump send packetPing err[%v] \n", err)
		//	ExitChanQueue <- 1
		//}
		case msg, ok := <-qs.MsgChan:
			if ok {
				//fmt.Printf("QueueServerMsgPump msg[%v] body[%v]\n", msg.ID, msg.Body)
				if msg.ID == 10012 {
					buf := new(bytes.Buffer)
					packer := binpacker.NewPacker(buf, binary.BigEndian)
					packer.PushString(string(msg.Body[:]))
					unpacker := binpacker.NewUnpacker(buf, binary.BigEndian)
					var flag byte
					if err := unpacker.FetchByte(&flag).Error(); err != nil {
						fmt.Printf("QueueServerMsgPump unpacker msgPing flag err[%v]\n", err)
						qs.ExitChan <- 1
					}
					//fmt.Printf("QueueServerMsgPump flag[%v]\n", flag)
					if flag == 0 {
						var queue, inQueue, time int32
						if err := unpacker.FetchInt32(&queue).FetchInt32(&inQueue).FetchInt32(&time).Error(); err != nil {
							fmt.Printf("QueueServerMsgPump unpacker msgPing err[%v]\n", err)
							qs.ExitChan <- 1
						}
						//fmt.Printf("QueueServerMsgPump  queue[%v] waitting No.[%v] time[%v]\n", queue, inQueue, time)
						fmt.Printf("[%v]当前排队总人数[%v]  前面等待人数[%v] 估计登录用时[%vs]\n", qs.Conn.LocalAddr(), queue, inQueue, time)
					} else {
						uuidlen := 36
						addrlen := uint64(msg.Len - 1 - uuidlen)
						var uuid, addr string
						if err := unpacker.FetchString(36, &uuid).FetchString(addrlen, &addr).Error(); err != nil {
							fmt.Printf("QueueServerMsgPump unpacker msgPing login err[%v]\n", err)
							qs.ExitChan <- 1
						}
						fmt.Printf("QueueServerMsgPump msgPing uuid[%v] addr[%v]\n", uuid, addr)

						// todo 拿着uuid登录loginServer

						var ls *LoginServer
						var err error
						if ls, err = NewLoginSer(); err != nil {
							fmt.Printf("new loginserver error[%v]\n", err)
							return
						}

						if err = ls.Connect2LoginServer(addr); err != nil {
							fmt.Printf("login server connect error[%v]\n", err)
							return
						}

						if err = ls.Start(); err != nil {
							fmt.Printf("login server start error[%v]\n", err)
						}

						// todo Check Server Run or Not and then send login packet
						ls.Run()

						if err = ls.SendLoginPakcetWithKey(uuid); err != nil {
							fmt.Printf("send login packet err[%v]\n", err)
							ls.Stop()
						}

						qs.ExitChan <- 1
					}
				}
			} else {
				fmt.Printf("QueueServerMsgPump from MsgChan not ok\n")
				qs.ExitChan <- 1
			}
		case <-qs.ExitChan:
			goto exit
		}
	}

exit:
	hbTicker.Stop()
	ppTicker.Stop()
	qs.Stop()
}

//func (qs *QueueServer) SendLoginPakcet() error {
//	// 向QueueServer发送登录信息
//	buf := new(bytes.Buffer)
//	packer := binpacker.NewPacker(buf, binary.BigEndian)
//	packer.PushByte(0x05)
//	packer.PushInt32(10013)
//	var flag byte
//	flag = '0'
//	accout := "account"
//	passwd := "passwd"
//	len := 1 + len(accout) + len(passwd)
//	packer.PushInt32((int32)(len))
//	packer.PushByte(flag)
//	packer.PushString(accout)
//	packer.PushString(passwd)
//	if err := packer.Error(); err != nil {
//		return fmt.Errorf("make msg err [%v]\n", err)
//	}
//
//	//fmt.Printf("client send c2slogin packet buf[%v] dataLen[%v]\n", buf.Bytes(), len)
//
//	if _, err := msgs.Send(qs, buf.Bytes()); err != nil {
//		return fmt.Errorf("send c2slogin packet err[%v] \n", err)
//	}
//	return nil
//}

func (qs *QueueServer) Connect2LoginServer(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("can not connect to server[%v]", addr)
	}

	qs.reader = bufio.NewReaderSize(conn, msgs.DefaultBufferSize)
	qs.writer = bufio.NewWriterSize(conn, msgs.DefaultBufferSize)
	qs.Conn = conn
	return nil
}

func (qs *QueueServer) Connect2QueueServer(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("can not connect to server[%v]", addr)
	}

	qs.Addr = addr
	qs.reader = bufio.NewReaderSize(conn, msgs.DefaultBufferSize)
	qs.writer = bufio.NewWriterSize(conn, msgs.DefaultBufferSize)
	qs.Conn = conn
	return nil
}

func NewQueueSer() (*QueueServer, error) {
	//conn, err := net.Dial("tcp", addr)
	//if err != nil {
	//	return nil, fmt.Errorf("can not connect to server[%v]", addr)
	//}
	//
	//qs := &QueueServer{
	//	reader: bufio.NewReaderSize(conn, msgs.DefaultBufferSize),
	//	writer: bufio.NewWriterSize(conn, msgs.DefaultBufferSize),
	//	Conn:   conn,
	//}

	qs := &QueueServer{}

	qs.MsgChan = make(chan *msgs.Message, 10000)
	qs.ExitChan = make(chan int, 1)
	qs.Pool.New = func() interface{} {
		return &msgs.Message{
			Timestamp: time.Now(),
		}
	}

	return qs, nil
}
