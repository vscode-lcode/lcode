package bash

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/SierraSoftworks/multicast/v2"
	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/bash/webdav"
	"github.com/vscode-lcode/lcode/v2/util/err0"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/vscode-lcode/lcode/v2/bash")

type Bash struct {
	clients *ttlcache.Cache[string, *webdav.Client]

	listener      net.Listener
	handler       []chan net.Conn
	handlerLocker sync.Locker
	connected     *multicast.Channel[*webdav.Client]

	IDGenerator IDGenerator
	VERSION     string
	ID          string
}

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
	sh.ID = fmt.Sprintf("lcode-hub@%s:%s", sh.VERSION, uuid.NewString())
	sh.listener = l
	for {
		conn := To1(l.Accept())
		go sh.serve(conn)
	}
}

func (sh *Bash) serve(conn net.Conn) (err error) {
	_, span := tracer.Start(context.Background(), "serve conn")
	defer span.End()
	defer err0.Record(&err, span)
	defer err2.Handle(&err, func() {
		conn.Close()
	})

	r := bufio.NewReader(conn)
	header, _ := To2(r.ReadLine())
	if len(header) == 0 {
		err = fmt.Errorf("unknown connect")
		return
	}
	switch v := string(header[0]); v {
	case "-":
		defer conn.Close()
		if string(header) == "-1" {
			io.WriteString(conn, sh.ID+"\n")
		}
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
