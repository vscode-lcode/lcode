package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/alessio/shellescape"
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vscode-lcode/lcode/v2/bash"
	"github.com/vscode-lcode/lcode/v2/bash/webdav"
	"github.com/vscode-lcode/lcode/v2/hub"
	"xorm.io/xorm"
)

var args struct {
	hello       string
	addr        string
	localdomain string
}

var VERSION = "dev"

func init() {
	flag.StringVar(&args.addr, "addr", "127.0.0.1:4349", "local-hub listen addr")
	flag.StringVar(&args.hello, "hello", "webdav://%s.lo.shynome.com:4349%s", "")
	flag.StringVar(&args.localdomain, "localdomain", ".lo.shynome.com", "")
}

func main() {
	flag.Parse()

	if err := hasRunning(args.addr); err == nil {
		return
	}

	lcodeDir := filepath.Join(To1(os.UserHomeDir()), ".config", "lcode/")
	To(os.MkdirAll(lcodeDir, os.ModePerm))
	db := To1(xorm.NewEngine("sqlite3", filepath.Join(lcodeDir, "lcode.db")))
	To(hub.Sync(db))

	l := To1(net.Listen("tcp", args.addr))
	fmt.Printf("lcode-hub is running on %s\n", args.addr)

	bash := bash.NewBash()
	bash.VERSION = VERSION
	go func() {
		var format = regexp.MustCompile(`^\d+-(.+)-(\d+)$`)
		for client := range bash.Connected() {
			go func(c *webdav.Client) {
				fmt.Println("client connected", c.ID)
				defer fmt.Println("client disconnected", c.ID)
				f := format.FindStringSubmatch(c.ID)
				if len(f) == 3 {
					id := fmt.Sprintf("%s-%s", f[2], f[1])
					hello := fmt.Sprintf(args.hello, id, c.PWD)
					c.Exec(fmt.Sprintf(">&2 echo lo: %s", shellescape.Quote(hello)))
				}
				<-c.Closed()
			}(client)
		}
	}()

	hub := hub.New(db, bash)
	hub.LocalDomain = args.localdomain
	To(hub.CleanClients())

	go http.Serve(net.Listener(bash), hub)
	To(bash.Serve(l))
}

func hasRunning(addr string) (err error) {
	defer err2.Handle(&err)
	client := http.Client{Timeout: 2 * time.Second}
	resp := To1(client.Get(fmt.Sprintf("http://%s/%s", addr, "version")))
	defer resp.Body.Close()
	r := bufio.NewReader(resp.Body)
	line, _ := To2(r.ReadLine())
	if v := string(line); v != "lcode-hub" {
		err = fmt.Errorf("expect lcode-hub, but got %s", v)
		return
	}
	version, _ := To2(r.ReadLine())
	fmt.Printf("lcode-hub already has running, version: %s. exit.\n", string(version))
	return
}
