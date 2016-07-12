package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/serverFramework/serverFramework/utils"

	"VODone/Tester/server"
)

type Tester struct {
	addr string
	wg   utils.WaitGroupWrapper

	ls server.LoginServer
	//qs server.QueueServer

	tps int
	qps int
}

func (t *Tester) Run(n int) {
	for i := 0; i < n; i++ {
		t.wg.Wrap(func() {
			t.startLoginServer()
		})
	}
}

func (t *Tester) startLoginServer() {
	var ls *server.LoginServer
	var err error
	if ls, err = server.NewLoginSer(); err != nil {
		fmt.Printf("new loginserver error[%v]\n", err)
		return
	}

	if err = ls.Connect2LoginServer(t.addr); err != nil {
		fmt.Printf("login server connect error[%v]\n", err)
		return
	}

	if err = ls.Start(); err != nil {
		fmt.Printf("login server start error[%v]\n", err)
	}

	// todo Check Server Run or Not and then send login packet
	ls.Run()

	if err = ls.SendLoginPakcet(); err != nil {
		fmt.Printf("send login packet err[%v]\n", err)
		ls.Stop()
	}
}

func (t *Tester) Stop() {
	t.ls.Stop()
	//t.qs.Stop()
	t.wg.Wait()
}

func (t *Tester) Wait() {
	t.wg.Wait()
}

func NewTester(addr string) *Tester {
	return &Tester{addr: addr}
}

func main() {
	maxClient := flag.Int("max", 5, "client")
	timer := flag.Int("time", 60, "the tester run time")

	fmt.Printf("client[%v] time[%v]\n", *maxClient, *timer)
	flag.Parse()

	tt := NewTester("127.0.0.1:60060")
	tt.Run(*maxClient)

	tt.Wait()

	// Handle SIGINT and SIGTERM.
	fmt.Printf("Print Ctrl+C to quit\n")
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println(<-ch)

	// Stop the service gracefully.
	tt.Stop()
}
