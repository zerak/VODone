package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"runtime"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zhuangsirui/binpacker"

	"bitbucket.org/serverFramework/serverFramework/core"
	"bitbucket.org/serverFramework/serverFramework/utils"

	_ "VODone/QueueServer/models"
	_ "VODone/QueueServer/msgs"
	"VODone/QueueServer/queue"
)

var wg utils.WaitGroupWrapper

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	core.ServerApp.Version("VODone/QueueServer")

	//// orm
	//core.SConfig.DBConf.User = "user"
	//core.SConfig.DBConf.PW = "pw"
	////core.SConfig.DBConf.Addr = "localhost"
	////core.SConfig.DBConf.Port = 3306
	////core.SConfig.DBConf.DB = "testDb"
	//str := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8",
	//	core.SConfig.DBConf.User, core.SConfig.DBConf.PW,
	//	core.SConfig.DBConf.Addr, core.SConfig.DBConf.Port,
	//	core.SConfig.DBConf.DB)
	//orm.RegisterDriver("mysql", orm.DRMySQL)
	//orm.RegisterDataBase("default", "mysql", str)
	//orm.SetMaxIdleConns("default", 30)
	//orm.SetMaxOpenConns("default", 30)

	addr := "127.0.0.1:60060"
	connect2LoginServer(addr)

	wg.Wrap(func() {
		serveQueue()
	})

	//core.Run()
	core.Run("127.0.0.1:60070")
	//core.Run("localhost")
	//core.Run(":60060")

	wg.Wait()
}

func serveQueue() {
	core.ServerLogger.Info("QueueServer start timer routine")
	for {
		select {
		case <- queue.NotifyChan:
			if queue.QueuedClients.Front() != nil && queue.AuthClients < queue.MaxClients {
				qc := queue.QueuedClients.Front().Value.(*queue.QueueClient)

				buf := new(bytes.Buffer)
				packer := binpacker.NewPacker(buf, binary.BigEndian)
				packer.PushByte(0x05)
				packer.PushInt32(10012)

				var flag byte
				uuid := qc.Session
				addr := "127.0.0.1:60060"
				len := 1 + len(uuid) + len(addr) // flag uuid addr
				flag = '1'
				packer.PushInt32((int32(len)))
				packer.PushByte(flag)
				packer.PushString(uuid)
				packer.PushString(addr)

				if _, err := notice(qc.Conn, buf.Bytes()); err != nil {
				}

				// TODO NOTE 向客户端发送重新登录消息,不管发送成功还是失败,都断开连接
				qc.Conn.Close()
			}
		}
	}

	defer func() {
		core.ServerLogger.Info("QueueServer timer routine exit\n")
	}()
}

func notice(conn net.Conn, by []byte) (n int, err error) {
	core.ServerLogger.Info("MsgLogin Queue server send to client notify relogin\n")
	if n, err := conn.Write(by); err != nil {
		return n, err
	}
	return n, nil
}
