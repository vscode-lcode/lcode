package bash

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/SierraSoftworks/multicast/v2"
	"github.com/alessio/shellescape"
	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/bash/webdav"
)

type Bash struct {
	clients *ttlcache.Cache[string, *webdav.Client]

	listener      net.Listener
	handler       []chan net.Conn
	handlerLocker sync.Locker
	connected     *multicast.Channel[*webdav.Client]

	IDGenerator IDGenerator
	VERSION     string
}

type ID interface {
	No() string
	NameSapce() string
	String() string
	Close() error
}
type IDGenerator func(id string, pwd string) (ID, error)

func NewBash() *Bash {
	return &Bash{
		clients:       ttlcache.New[string, *webdav.Client](),
		handler:       []chan net.Conn{},
		handlerLocker: &sync.Mutex{},
		connected:     multicast.New[*webdav.Client](),
		VERSION:       "dev",
	}
}

func (sh *Bash) Serve(l net.Listener) (err error) {
	defer err2.Handle(&err)
	if sh.IDGenerator == nil {
		return fmt.Errorf("IDGenerator is required")
	}
	sh.listener = l
	for {
		conn := To1(l.Accept())
		go sh.serve(conn)
	}
}

func (sh *Bash) serve(conn net.Conn) (err error) {
	defer err2.Handle(&err, func() {
		conn.Close()
		fmt.Println("serve err:", err)
	})

	r := bufio.NewReader(conn)
	header, _ := To2(r.ReadLine())
	if len(header) == 0 {
		err = fmt.Errorf("unknown connect")
		return
	}
	switch v := string(header[0]); v {
	case "0":
		sh.Connect(r, conn)
	case "1":
		var ids = strings.SplitN(string(header[1:]), ":", 2)
		if len(ids) != 2 {
			err = fmt.Errorf("task request need include cid and tid")
			return
		}
		cid, tid := ids[1], ids[0]
		client := sh.clients.Get(cid)
		if client == nil {
			err = fmt.Errorf("client is gone")
			return
		}
		bc := NewBashConn(r, conn)
		To(client.Value().Recive(tid, bc))
	case "2":
		sh.BindStderr(string(header), r, conn)
	default:
		reader := io.MultiReader(bytes.NewReader(header), bytes.NewReader([]byte("\n")), r)
		bc := NewBashConn(reader, conn)
		sh.toHandler(bc)
	}
	return
}

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

	To(c.Open(r, sh.VERSION))
	defer c.Close()
	idRaw := strings.ToLower(c.ID)

	tmpID := uuid.NewString()
	c.ID = tmpID
	sh.clients.Set(tmpID, c, ttlcache.NoTTL) // we can exec cmd after set client
	defer sh.clients.Delete(tmpID)

	id := To1(sh.IDGenerator(idRaw, c.PWD))
	defer id.Close()
	if !strings.HasPrefix(idRaw, id.No()+"-") && sh.VERSION != "dev" { // no 不一致时覆盖已有的 no
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

func (sh *Bash) BindStderr(v string, r *bufio.Reader, conn net.Conn) (err error) {
	defer err2.Handle(&err)
	defer conn.Close()
	var filters = []string{"lo: ", "1>"}
	defer fmt.Println("err ch closed")
	switch v {
	case "2":
		fallthrough
	case "22":
		if v == "2" {
			filters = filters[:1]
		}
		for {
			b, _ := To2(r.ReadLine())
			go func(line string) {
				for _, prefix := range filters {
					if strings.HasPrefix(line, prefix) {
						io.WriteString(conn, line+"\n")
					}
				}
			}(string(b))
		}
	default:
		io.Copy(conn, r)
	}
	return
}
