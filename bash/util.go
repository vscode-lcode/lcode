package bash

import (
	"bytes"
	"io"
	"net"
	"os/exec"

	. "github.com/lainio/err2/try"
)

func StartTestClient(addr net.Addr) (cmd *exec.Cmd) {
	cmd = exec.Command("bash", "-i")
	conn := To1(net.Dial("tcp", addr.String()))
	cmd.Stdin = conn
	cmd.Stdout = conn
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	// cmd.Stderr = os.Stderr
	To1(io.WriteString(conn, "0\n"))
	go cmd.Start()
	return
}
