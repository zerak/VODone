package main

import (
	//"fmt"
	"runtime"
	"sync/atomic"

	//"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"

	"bitbucket.org/serverFramework/serverFramework/core"
	"bitbucket.org/serverFramework/serverFramework/utils"

	_ "VODone/QueueServer/models"
	_ "VODone/QueueServer/msgs"
)

var wg utils.WaitGroupWrapper
var MaxClients int32
var CurClients int32
var InterVal int32

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

	//core.Run()
	core.Run("127.0.0.1:60070")
	//core.Run("localhost")
	//core.Run(":60060")

	//wg.Wait()
}

func SetMaxClients(n int32) {
	atomic.StoreInt32(&MaxClients, n)
}

func GetMaxClients() int32 {
	return atomic.LoadInt32(&MaxClients)
}

func SetCurClients(n int32) {
	atomic.StoreInt32(&CurClients, n)
}

func GetCurClients() int32 {
	return atomic.LoadInt32(&CurClients)
}

func SetInterVal(n int32) {
	atomic.StoreInt32(&InterVal, n)
}

func GetInterVal() int32 {
	return atomic.LoadInt32(&InterVal)
}
