package bash

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/bash/webdav"
)

type ID interface {
	No() string
	NameSapce() string
	String() string
	Close() error
}
type LcodeClient interface {
	RawID() string
	PWD() string
	Targets() []string
}
type IDGenerator func(client LcodeClient) (ID, error)

func (sh *Bash) Connect(r *bufio.Reader, conn net.Conn) (err error) {
	defer conn.Close()
	defer err2.Handle(&err, func() {
		if errors.Is(err, io.EOF) {
			return
		}
		if errors.Is(err, webdav.ErrNeedPrint) {
			cmd := fmt.Sprintf(">&2 echo lo: %s\n", shellescape.Quote(err.Error()))
			io.WriteString(conn, cmd)
			return
		}
		if errors.Is(err, webdav.ErrPrintHelp) {
			return
		}
		fmt.Println("client connect", err)
	})

	io.WriteString(conn, "export PS1=''\n")

	c := webdav.NewClient(conn)

	To(c.Open(r, sh.VERSION, sh.ID))
	defer c.Close()
	lc := LcodeClientWrapper{c, c.ID}

	tmpID := uuid.NewString()
	c.ID = tmpID
	sh.clients.Set(tmpID, c, ttlcache.NoTTL) // we can exec cmd after set client
	defer sh.clients.Delete(tmpID)

	id := To1(sh.IDGenerator(lc))
	defer id.Close()
	if !strings.HasPrefix(lc.RawID(), id.No()+"-") && sh.VERSION != "dev" { // no 不一致时覆盖已有的 no
		// 开发环境不更新原有的 no
		To(c.StoreID(id.No()))
	}
	sh.clients.Delete(tmpID)
	c.ID = id.String()
	sh.clients.Set(c.ID, c, ttlcache.NoTTL)
	defer sh.clients.Delete(c.ID)

	sh.connected.C <- c

	for {
		_, _, err = r.ReadLine()
		if err != nil {
			break
		}
		// fmt.Println("line:", string(line))
	}
	return
}

type LcodeClientWrapper struct {
	*webdav.Client
	rawID string
}

var _ LcodeClient = (*LcodeClientWrapper)(nil)

func (lc LcodeClientWrapper) RawID() string { return lc.rawID }
func (lc LcodeClientWrapper) PWD() string   { return lc.Client.PWD }
