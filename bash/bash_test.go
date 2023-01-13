package bash

import (
	"fmt"
	"net"

	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/bash/webdav"
)

var bash = NewBash()
var client *webdav.Client

type TestClientID string

var testClientIDNo = 0

func (id TestClientID) No() string        { return string(id) }
func (id TestClientID) NameSapce() string { return "test" }
func (id TestClientID) String() string {
	return fmt.Sprintf("%s-%s", id.NameSapce(), id.No())
}
func (id TestClientID) Close() error { return nil }

var l = To1(net.Listen("tcp", ":0"))

func init() {

	bash.IDGenerator = func(c LcodeClient) (ID, error) {
		testClientIDNo++
		no := testClientIDNo
		return TestClientID(fmt.Sprintf("%d", no)), nil
	}

	go func() {
		To(bash.Serve(l))
	}()

	connected := bash.Connected()
	if true {
		StartTestClient(l.Addr())
	}

	client = <-connected
}
