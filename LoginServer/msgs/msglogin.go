package msgs

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/zhuangsirui/binpacker"

	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"
)

type MsgLogin struct {
}

func (m *MsgLogin) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("MsgLogin cid[%v] msg login msgId[%v] msgBody[%v] auth[%v]Allo[%v]", client.GetID(), msg.ID, msg.Body, ServerApp.GetAuthClients(), ServerApp.GetAllowClients())
	ServerLogger.Info("MsgLogin auth? ", client.GetAuth())
	// A TODO
	// 检查客户端发送包是否有排队session标识
	// 如果是之前排过队,并且当前服务器未超过最大客户端数量,则客户端直接进入登录流程

	// 1,检查是否需要排队
	bufbody := new(bytes.Buffer)
	pp := binpacker.NewPacker(bufbody, binary.BigEndian)
	pp.PushString(string(msg.Body[:]))
	unpackbody := binpacker.NewUnpacker(bufbody, binary.BigEndian)
	var flag byte
	if err := unpackbody.FetchByte(&flag).Error(); err != nil {
		ServerLogger.Error("MsgLogin unpacker flag err", err)
		client.Exit()
		return
	}

	ServerLogger.Info("MsgLogin unpacker flag[%v]", flag)

	// 首次登录,未经过排队
	if flag == 48 {
		strlen := uint64(msg.Len - 1)
		ServerLogger.Info("MsgLogin account/passwd len[%v]", strlen, bufbody.Bytes())
		var str string
		if err := unpackbody.FetchString(strlen, &str).Error(); err != nil {
			ServerLogger.Error("MsgLogin unpacker account/passwd err", err)
			client.Exit()
			return
		}

		loginbuf := new(bytes.Buffer)
		packer := binpacker.NewPacker(loginbuf, binary.BigEndian)
		packer.PushByte(0x05)
		packer.PushInt32(10014)

		succ := false // 是否排队
		if ServerApp.GetAuthClients() >= ServerApp.GetAllowClients() {
			succ = true
		}
		if succ {
			// 返回登录失败包,客户端重定向到QueueServer
			var flag byte
			flag = '0'
			addr := "127.0.0.1:60070"
			len := 1 + len(addr)
			packer.PushInt32((int32)(len))
			packer.PushByte(flag)
			packer.PushString(addr)

			client.SetAuth(false) // 设置未认证标识
			ServerLogger.Info("MsgLogin login failed redirect to queue server")
		} else {
			// todo
			// 2,解析登录包,帐号密码验证成功后,返回登录成功包
			var flag byte
			flag = '1'           // 1 byte
			id := client.GetID() // 8 byte
			name := client.String()
			len := 1 + 8 + len(name)
			packer.PushInt32((int32(len)))
			packer.PushByte(flag)
			packer.PushInt64(id).PushString(name)

			// todo
			// 3,异步登录成功后,服务器更新当前登录人数及客户端认证状态
			ServerApp.AuthClient(true)
			client.SetAuth(true)
			ServerLogger.Info("MsgLogin login succ")
		}

		if err := packer.Error(); err != nil {
			ServerLogger.Info("MsgLogin pack msg err [%v]\n", err)
			client.Exit()
			return
		}

		ServerLogger.Info("buf[%x]", loginbuf.Bytes())

		if _, err := p.Send(client, loginbuf.Bytes()); err != nil {
			err = fmt.Errorf("failed to send response ->%s", err)
			ServerLogger.Error("MsgLogin [%v]\n", err)
			client.Exit()
		}
	} else {
		uuidLen := 36
		var uuid string
		if err := unpackbody.FetchString((uint64)(uuidLen), &uuid).Error(); err != nil {
			ServerLogger.Error("MsgLogin unpack uuid err")
			client.Exit()
			return
		}
		// todo Verify uuid
		ServerLogger.Info("MsgLogin uuid[%v]", uuid)

		loginbuf := new(bytes.Buffer)
		packer := binpacker.NewPacker(loginbuf, binary.BigEndian)
		packer.PushByte(0x05)
		packer.PushInt32(10014)

		succ := false // 已认证是否大于服务器最大承载量
		if ServerApp.GetAuthClients() >= ServerApp.GetMaxClients() {
			succ = true
			ServerLogger.Info("MsgLogin uth[%v] big than max[%v] relogin failed", ServerApp.GetAuthClients(), ServerApp.GetMaxClients())
		}
		if succ {
			// 返回登录失败包,客户端重定向到QueueServer
			var flag byte
			flag = '0'
			addr := "127.0.0.1:60070"
			len := 1 + len(addr)
			packer.PushInt32((int32)(len))
			packer.PushByte(flag)
			packer.PushString(addr)

			client.SetAuth(false) // 设置未认证标识
			ServerLogger.Info("MsgLogin login failed redirect to queue server")
		} else {
			// todo
			// 2,解析登录包,帐号密码验证成功后,返回登录成功包
			var flag byte
			flag = '1'           // 1 byte
			id := client.GetID() // 8 byte
			name := client.String()
			len := 1 + 8 + len(name)
			packer.PushInt32((int32(len)))
			packer.PushByte(flag)
			packer.PushInt64(id).PushString(name)

			// todo
			// 3,异步登录成功后,服务器更新当前登录人数及客户端认证状态
			ServerApp.AuthClient(true)
			client.SetAuth(true)
			ServerLogger.Info("MsgLogin login with uuid succ")
		}

		if err := packer.Error(); err != nil {
			ServerLogger.Info("MsgLogin pack msg err [%v]\n", err)
			client.Exit()
			return
		}

		ServerLogger.Info("buf[%x]", loginbuf.Bytes())

		if _, err := p.Send(client, loginbuf.Bytes()); err != nil {
			err = fmt.Errorf("failed to send response ->%s", err)
			ServerLogger.Error("MsgLogin [%v]\n", err)
			client.Exit()
		}
	}

	ServerLogger.Info("MsgLogin max client[%v] authClients[%v]", ServerApp.GetMaxClients(), ServerApp.GetAuthClients())
}
