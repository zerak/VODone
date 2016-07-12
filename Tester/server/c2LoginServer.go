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

type LoginServer struct {
	queue service.ServiceQueue

	reader    *bufio.Reader
	writer    *bufio.Writer
	writeLock sync.RWMutex

	MsgChan  chan *msgs.Message
	ExitChan chan int
	Pool     sync.Pool
	Conn     net.Conn
	IsConn   bool
	Addr     string

	wg utils.WaitGroupWrapper
}

func (ls *LoginServer) SetQueueServerPtr(queue service.ServiceQueue) {
	ls.queue = queue
}

func (ls *LoginServer) GetReader() *bufio.Reader {
	return ls.reader
}

func (ls *LoginServer) GetWriter() *bufio.Writer {
	return ls.writer
}

func (ls *LoginServer) GetConn() net.Conn {
	return ls.Conn
}

func (ls *LoginServer) GetLock() sync.RWMutex {
	return ls.writeLock
}

func (ls *LoginServer) Start() error {
	if _, err := msgs.Send(ls, []byte("  V1")); err != nil {
		return fmt.Errorf("send protocol [  v1]err [%v]\n", err)
	}

	return nil
}

func (ls *LoginServer) Run() {
	ls.wg.Wrap(func() {
		ls.serverLoop()
	})
}

func (ls *LoginServer) Stop() {
	if ls.IsConn {
		ls.Conn.Close()
		close(ls.ExitChan)
	}
	ls.wg.Wait()
}

func (ls *LoginServer) serverLoop() {
	var err error
	var header byte
	var cmd uint32
	var length uint32

	msgPumpStartedChan := make(chan bool)
	go ls.msgPump(msgPumpStartedChan)
	<-msgPumpStartedChan

	fmt.Printf("LoginServerLoop remoteAddr[%v] localAddr[%v]\n", ls.Conn.RemoteAddr(), ls.Conn.LocalAddr())

	buf := make([]byte, msgs.ProtocolHeaderLen)
	for {
		_, err = io.ReadFull(ls.reader, buf)
		if err != nil {
			fmt.Printf("LoginServerLoop read head from remote[%v] err->%v buffed->%v\n", ls.Conn.RemoteAddr(), err, ls.reader.Buffered())
			//ExitChan <- 1
			break
		}

		// header
		header = buf[0]
		if header != 0x05 {
			err = fmt.Errorf("LoginServerLoop header[%s] err", header)
			//ExitChan <- 1
			break
		}

		// cmd
		cmd = binary.BigEndian.Uint32(buf[1:5])

		// length
		length = binary.BigEndian.Uint32(buf[5:9])

		// data
		data := make([]byte, length)
		_, err = io.ReadFull(ls.reader, data)
		if err != nil {
			fmt.Printf("LoginServerLoop read data from client[%v] err->%v buffed->%v", ls.Conn.RemoteAddr(), err, ls.reader.Buffered())
			//ExitChan <- 1
			break
		}

		//fmt.Printf("LoginServerLoop cid[%v] header[%v] cmd[%v] len[%d] data[%x]\n", client.LocalAddr().String(), header, cmd, length, data)

		// new msg
		//msg := Pool.Get().(*msgs.Message)
		//msg := &msgs.Message{ID:(int32)(cmd),Body:data,Conn:client}
		var msg msgs.Message
		msg.ID = int(cmd)
		msg.Body = data
		msg.Len = (int)(length)
		msg.Conn = ls.Conn

		ls.MsgChan <- &msg
	}

	defer func() {
		ls.Conn.Close()
		//ls.ExitChan <- 1
		fmt.Printf("LoginServerLoop cid[%v] exit\n", ls.Conn.LocalAddr().String())
	}()
}

func (ls *LoginServer) msgPump(startedChan chan bool) {
	close(startedChan)

	hbTickerLogin := time.NewTicker(msgs.C2LoginServerHB)
	hbChanLogin := hbTickerLogin.C
	quit := false
	for {
		select {
		case <-hbChanLogin:
			buf := new(bytes.Buffer)
			packer := binpacker.NewPacker(buf, binary.BigEndian)
			packer.PushByte(0x05)
			packer.PushInt32(10010)
			packer.PushInt32(0)
			if err := packer.Error(); err != nil {
				fmt.Printf("LoginServerMsgPump make msg err [%v]\n", err)
				ls.ExitChan <- 1
			}

			//fmt.Printf("LoginServerMsgPump heartbeat cid[%v] buf[%x] \n", client.LocalAddr().String(), buf.Bytes())

			if n, err := msgs.Send(ls, buf.Bytes()); err != nil || n != 9 {
				fmt.Printf("LoginServerMsgPump send heartbeat packet err[%v] \n", err)
				ls.ExitChan <- 1
			} else {
				//fmt.Printf("msg hb len[%v]\n", n)
			}
		case msg, ok := <-ls.MsgChan:
			if ok {
				//fmt.Printf("LoginServerMsgPump cid[%v] msgChan msg[%v] body[%v]\n", client.LocalAddr().String(), msg.ID, msg.Body)
				if msg.ID == 10014 {
					buf := new(bytes.Buffer)
					packer := binpacker.NewPacker(buf, binary.BigEndian)
					packer.PushString(string(msg.Body[:]))
					unpacker := binpacker.NewUnpacker(buf, binary.BigEndian)

					var flag byte
					if err := unpacker.FetchByte(&flag).Error(); err != nil {
						fmt.Printf("LoginServerMsgPump unpacker err[%v]\n", err)
						ls.ExitChan <- 1
					}

					//fmt.Printf("LoginServerMsgPump cid[%v] flag[%v]\n", client.LocalAddr().String(), flag)
					if flag == 48 {
						//todo
						// login server return err, and connect to queue server
						var addr string
						len := uint64(msg.Len - 1)
						if err := unpacker.FetchString(len, &addr).Error(); err != nil {
							fmt.Printf("LoginServerMsgPump login failed and get queue server addr err\n")
							ls.ExitChan <- 1
						}
						fmt.Printf("LoginServerMsgPump login failed and redirect to queue server[%v]\n", addr)

						var qs *QueueServer
						var err error
						if qs, err = NewQueueSer(); err != nil {
							fmt.Printf("LoginServerMsgPump new queueserver error[%v]\n", err)
							ls.ExitChan <- 1
						}

						if err = qs.Connect2QueueServer(addr); err != nil {
							fmt.Printf("login server connect error[%v]\n", err)
							ls.ExitChan <- 1
						}

						if err = qs.Start(); err != nil {
							fmt.Printf("login server start error[%v]\n", err)
							ls.ExitChan <- 1
						}

						// todo Check Server Run or Not and then send login packet
						qs.Run()

						ls.ExitChan <- 1
					} else {
						var uid int64
						var name string
						len := uint64(msg.Len - 1 - 8)
						if err := unpacker.FetchInt64(&uid).FetchString(len, &name).Error(); err != nil {
							fmt.Printf("LoginServerMsgPump login success but unpack failed [%v]\n", err)
							ls.ExitChan <- 1
						}
						fmt.Printf("LoginServerMsgPump login success uid[%v] name[%v]\n", uid, name)
					}
				}
			} else {
				fmt.Printf("LoginServerMsgPump from MsgChan not ok\n")
				ls.ExitChan <- 1
			}
		case <-ls.ExitChan:
			fmt.Printf("LoginServerMsgPump exitChan recv EXIT\n")
			quit = true
		}
		if quit {
			break
		}
	}

	defer func() {
		hbTickerLogin.Stop()
		//ls.Conn.Close()
		//close(ls.ExitChan)
		ls.Stop()
		fmt.Printf("LoginServerMsgPump loginserver msg pump stop addr[%v]\n", ls.Conn.LocalAddr().String())
	}()
}

// TODO package msgs will do this
func (ls *LoginServer) SendLoginPakcet() error {
	// 向LoginServer发送登录信息
	buf := new(bytes.Buffer)
	packer := binpacker.NewPacker(buf, binary.BigEndian)
	packer.PushByte(0x05)
	packer.PushInt32(10013)
	var flag byte
	flag = '0'
	accout := "account"
	passwd := "passwd"
	len := 1 + len(accout) + len(passwd)
	packer.PushInt32((int32)(len))
	packer.PushByte(flag)
	packer.PushString(accout)
	packer.PushString(passwd)
	if err := packer.Error(); err != nil {
		return fmt.Errorf("make msg err [%v]\n", err)
	}

	//fmt.Printf("client send c2slogin packet buf[%v] dataLen[%v]\n", buf.Bytes(), len)

	if _, err := msgs.Send(ls, buf.Bytes()); err != nil {
		return fmt.Errorf("send c2slogin packet err[%v] \n", err)
	}
	return nil
}

func (ls *LoginServer) SendLoginPakcetWithKey(uuid string) error {
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
		return fmt.Errorf("make msg err [%v]\n", err)
	}

	//fmt.Printf("client reconnect to loginserver c2slogin packet buf[%x] dataLen[%v]\n", buf.Bytes(), len)

	if _, err := msgs.Send(ls, buf.Bytes()); err != nil {
		return fmt.Errorf("send c2slogin packet err[%v] \n", err)
	}
	return nil
}

func (ls *LoginServer) Connect2LoginServer(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("can not connect to server[%v]", addr)
	}

	ls.Addr = addr
	ls.reader = bufio.NewReaderSize(conn, msgs.DefaultBufferSize)
	ls.writer = bufio.NewWriterSize(conn, msgs.DefaultBufferSize)
	ls.Conn = conn
	ls.IsConn = true
	return nil
}

func (ls *LoginServer) Connect2QueueServer(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("can not connect to server[%v]", addr)
	}

	ls.reader = bufio.NewReaderSize(conn, msgs.DefaultBufferSize)
	ls.writer = bufio.NewWriterSize(conn, msgs.DefaultBufferSize)
	ls.Conn = conn
	return nil
}

func (ls *LoginServer) Reset(lsNew *LoginServer) {
	ls = lsNew
}

func NewLoginSer() (*LoginServer, error) {
	//conn, err := net.Dial("tcp", addr)
	//if err != nil {
	//	return nil, fmt.Errorf("can not connect to server[%v]", addr)
	//}
	//
	//ls := &LoginServer{
	//	reader: bufio.NewReaderSize(conn, msgs.DefaultBufferSize),
	//	writer: bufio.NewWriterSize(conn, msgs.DefaultBufferSize),
	//	Conn:   conn,
	//}

	ls := &LoginServer{}
	ls.MsgChan = make(chan *msgs.Message, 10000)
	ls.ExitChan = make(chan int, 1)
	ls.Pool.New = func() interface{} {
		return &msgs.Message{
			Timestamp: time.Now(),
		}
	}
	ls.IsConn = false

	return ls, nil
}
