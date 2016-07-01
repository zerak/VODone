package msgs

import (
	"fmt"
	"strconv"

	"github.com/astaxie/beego/orm"

	. "bitbucket.org/serverFramework/serverFramework/client"
	. "bitbucket.org/serverFramework/serverFramework/core"
	. "bitbucket.org/serverFramework/serverFramework/protocol"

	. "VODone/LoginServer/models"
)

type MsgLogin struct {
}

func init() {
	RegisterMsg(strconv.Itoa(10011), &MsgLogin{})
}

func (m *MsgLogin) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("cid[%v] msg login", client.GetID())

	o := orm.NewOrm()
	o.Using("default")
	//u := new(Mtable)
	u := new(Table)
	u.Time = msg.Timestamp
	_, err := o.Insert(u)
	if err != nil {
		ServerLogger.Error("login error %v", err)
		_, err := p.Send(client, []byte("s2c login error"))
		if err != nil {
			err = fmt.Errorf("failed to send response ->%s", err)
			client.Exit()
		}
	} else {
		e := o.Read(u)
		if e != nil {
			ServerLogger.Info("read", u.ID, u.Time)
		}

		_, err := p.Send(client, []byte("s2c login succ"))
		if err != nil {
			err = fmt.Errorf("failed to send response ->%s", err)
			client.Exit()
		}
	}

}

