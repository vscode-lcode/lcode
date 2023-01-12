package hub

import (
	"net"
	"time"

	. "github.com/lainio/err2/try"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vscode-lcode/lcode/v2/bash"
	"xorm.io/xorm"
)

var hub *Hub

func init() {
	eg := To1(xorm.NewEngine("sqlite3", "./lcode.db"))
	To(Sync(eg))

	l := To1(net.Listen("tcp", ":0"))
	sh := bash.NewBash()

	hub = New(eg, sh)

	go func() {
		To(sh.Serve(l))
	}()
	time.Sleep(time.Second)

	bash.StartTestClient(l.Addr())
	time.Sleep(time.Second)

}
