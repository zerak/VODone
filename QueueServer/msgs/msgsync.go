package msgs

//
//import (
//	//"fmt"
//	//"bytes"
//	//"encoding/binary"
//	//
//	//"github.com/zhuangsirui/binpacker"
//
//	. "bitbucket.org/serverFramework/serverFramework/client"
//	. "bitbucket.org/serverFramework/serverFramework/core"
//	. "bitbucket.org/serverFramework/serverFramework/protocol"
//)
//
//type MsgSync struct {
//}
//
//func (m *MsgSync) ProcessMsg(p Protocol, client Client, msg *Message) {
//	ServerLogger.Info("cid[%v] msg sync", client.GetID())
//
//	//buf := new(bytes.Buffer)
//	//packer := binpacker.NewPacker(buf, binary.BigEndian)
//	//
//	//packer.PushByte(0x05)
//	//packer.PushInt32(60001)
//	//Max := ServerApp.GetMaxClients()
//	//Cur := ServerApp.GetCurrentClients()
//	//Time := 10
//	////len := len(int32(Max)) + len(int32(Cur)) + len(int32(Time))
//	//len := 4 + 4 + 4
//	//packer.PushInt32((int32)(len))
//	//
//	//packer.PushUint32(uint32(Max))
//	//packer.PushUint32(uint32(Cur))
//	//packer.PushUint32(uint32(Time))
//	//
//	//if err := packer.Error(); err != nil {
//	//	ServerLogger.Info("make msg err [%v]", err)
//	//	client.Exit()
//	//}
//	//
//	//ServerLogger.Info("msgSync s2c buf[%x] len[%v]", buf.Bytes(), len)
//	//if _, err := p.Send(client, buf.Bytes()); err != nil {
//	//	err = fmt.Errorf("failed to send response ->%s", err)
//	//	client.Exit()
//	//}
//}
