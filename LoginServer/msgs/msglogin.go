package msgs

import (
	"bytes"
	"encoding/binary"
	"fmt"

	//"github.com/astaxie/beego/orm"
	"github.com/zhuangsirui/binpacker"

	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"
)

type MsgLogin struct {
}

func (m *MsgLogin) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("MsgLogin cid[%v] msg login", client.GetID())
	ServerLogger.Info("MsgLogin msgId[%v] msgBody[%v]", msg.ID, msg.Body)

	ServerLogger.Info("max client[%v] curClients[%v]", ServerApp.GetMaxClients(), ServerApp.GetCurrentClients())

	// todo
	// 解析登录包后,未做帐号密码验证
	buf := new(bytes.Buffer)
	packer := binpacker.NewPacker(buf, binary.BigEndian)
	packer.PushByte(0x05)
	packer.PushInt32(10014)
	if true {
		// 返回登录失败包
		var flag byte
		flag = '0'
		addr := "127.0.0.1:60070"
		len := 1 + len(addr)
		packer.PushInt32((int32)(len))
		packer.PushByte(flag)
		packer.PushString(addr)
	}
	//if true {
	//	// 返回登录成功包
	//	flag := '1'
	//	id := client.GetID()
	//	name := client.GetID()
	//	len := len(flag) + len(id) + len(name)
	//	packer.PushInt32(len)
	//	packer.PushByte(flag)
	//	packer.PushString(id).PushString(name)
	//
	//}
	if err := packer.Error(); err != nil {
		fmt.Printf("make msg err [%v]\n", err)
		panic(err)
	}

	ServerLogger.Info("buf[%x]", buf.Bytes())
	if _, err := p.Send(client, buf.Bytes()); err != nil {
		err = fmt.Errorf("failed to send response ->%s", err)
		client.Exit()
	}
}
