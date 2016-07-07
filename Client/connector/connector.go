package connector

//import (
//	"net"
//	"fmt"
//	"bufio"
//	"time"
//	"bytes"
//	"encoding/binary"
//
//	"github.com/zhuangsirui/binpacker"
//
//	"VODone/Client/login"
//	"VODone/Client/queue"
//	"VODone/Client/msgs"
//)
//
//func Connect2QueueServer(addr string) net.Conn {
//	fmt.Printf("connect2QueueServer [%v]\n", addr)
//	conn, err := net.Dial("tcp", addr)
//	if err != nil {
//		panic(err)
//	}
//
//	queue.ReaderQueue = bufio.NewReaderSize(conn, msgs.DefaultBufferSize)
//	queue.WriterQueue = bufio.NewWriterSize(conn, msgs.DefaultBufferSize)
//
//	queue.StartQueueServerLoop(conn)
//
//	return conn
//}
//
//func Connect2LoginServer(addr string) net.Conn {
//	conn, err := net.Dial("tcp", addr)
//	if err != nil {
//		panic(err)
//	}
//
//	login.ReaderLogin = bufio.NewReaderSize(conn, msgs.DefaultBufferSize)
//	login.WriterLogin = bufio.NewWriterSize(conn, msgs.DefaultBufferSize)
//
//	login.StartLoginServerLoop(conn)
//	return conn
//}
//
//func SendLoginPakcet(conn net.Conn, interVal time.Duration) {
//	// todo
//	// 向LoginServer发送登录信息
//	ticker := time.NewTicker(time.Second * interVal)
//	for _ = range ticker.C {
//		buf := new(bytes.Buffer)
//		packer := binpacker.NewPacker(buf, binary.BigEndian)
//		packer.PushByte(0x05)
//		packer.PushInt32(10013)
//		accout := "account"
//		passwd := "passwd"
//		len := len(accout) + len(passwd)
//		packer.PushInt32((int32)(len))
//		packer.PushString(accout).PushString(passwd)
//		if err := packer.Error(); err != nil {
//			fmt.Printf("make msg err [%v]\n", err)
//			panic(err)
//		}
//
//		fmt.Printf("client send c2slogin packet buf[%x] dataLen[%v]\n", buf.Bytes(), len)
//
//		if _, err := login.Send2Login(conn, buf.Bytes()); err != nil {
//			fmt.Printf("send c2slogin packet err[%v] \n", err)
//			panic(err)
//		}
//
//		ticker.Stop()
//	}
//}
