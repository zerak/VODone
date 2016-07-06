package main

import (
	"fmt"
	"flag"

	"bitbucket.org/serverFramework/serverFramework/utils"

	"VODone/Client/login"
)

var addr = "127.0.0.1:60060"

var wg utils.WaitGroupWrapper

func main() {
	maxClient := flag.Int("max", 100, "client")
	tt := flag.Int("time",60,"the tester run time")

	fmt.Printf("client[%v] time[%v]\n", *maxClient, *tt)

	flag.Parse()

	for i := 0; i < *maxClient; i++ {
		wg.Wrap(func() {
			conn := login.Connect2LoginServer(addr)
			login.SendLoginPakcet(conn,2)
			//time.Sleep(time.Microsecond)
		})
	}

	wg.Wait()
}
