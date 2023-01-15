package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/alessio/shellescape"
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/bash"
	"github.com/vscode-lcode/lcode/v2/bash/webdav"
	"github.com/vscode-lcode/lcode/v2/hub"
	"github.com/vscode-lcode/lcode/v2/util/err0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

var args struct {
	hello       string
	addr        string
	localdomain string
	logLv       string
}

var VERSION = "dev"
var f = flag.NewFlagSet("lcode-hub@"+VERSION, flag.ExitOnError)
var defaultLogLv = "11"

func init() {
	f.StringVar(&args.addr, "addr", "127.0.0.1:4349", "local-hub listen addr")
	f.StringVar(&args.hello, "hello", "webdav://{{.host}}.lo.shynome.com:4349{{.path}}", "")
	f.StringVar(&args.localdomain, "localdomain", ".lo.shynome.com", "")
	f.StringVar(&args.logLv, "log", defaultLogLv, "日志输出等级: 0 - 不输出; 1 - Error; 11 - Info; 111 - Debug")
}

var tracer = otel.Tracer("lcode-hub")

func main() {
	f.Parse(os.Args[1:])

	loglv := To1(strconv.ParseInt(args.logLv, 2, 64))
	tp := newTracerProvider(LogLevel(loglv))
	defer tp.Shutdown(context.Background())

	if err := hasRunning(args.addr); err == nil {
		return
	}

	lcodeDir := filepath.Join(To1(os.UserHomeDir()), ".config", "lcode/")
	To(os.MkdirAll(lcodeDir, os.ModePerm))
	db := To1(xorm.NewEngine("sqlite", filepath.Join(lcodeDir, "lcode.db")))
	To(hub.Sync(db))

	l := To1(net.Listen("tcp", args.addr))
	defer l.Close()

	_, span := tracer.Start(
		err0.WithStatus(nil, codes.Ok, ""),
		"lcode-hub is running",
		trace.WithAttributes(attribute.String("addr", args.addr)),
	)
	defer span.End()

	bash := bash.NewBash()
	bash.VERSION = VERSION
	go func() {
		var format = regexp.MustCompile(`^\d+-(.+)-(\d+)$`)
		helloTpl := To1(template.New("hello").Parse(args.hello + "\n"))
		for client := range bash.Connected() {
			go func(c *webdav.Client) {
				_, span := tracer.Start(
					err0.WithStatus(err0.KeepEndOutput(nil), codes.Ok, ""),
					"client connected",
					trace.WithAttributes(attribute.String("id", c.ID)),
				)
				defer span.End()

				f := format.FindStringSubmatch(c.ID)
				if len(f) != 3 {
					return
				}
				id := fmt.Sprintf("%s-%s", f[2], f[1])
				noEditTargets := true
				var output bytes.Buffer
				for _, t := range c.Targets() {
					switch {
					case strings.HasPrefix(t, "/dev/null"):
						t = strings.TrimPrefix(t, "/dev/null")
						fmt.Fprintf(&output, "this target is not exists: %s", t)
					case strings.HasPrefix(t, "/dev/err"):
						t = strings.TrimPrefix(t, "/dev/err")
						fmt.Fprintf(&output, "this target cannot be opened: %s", t)
					default:
						noEditTargets = false
						helloTpl.Execute(&output, map[string]string{
							"host": id,
							"path": t,
						})
					}
				}
				hello := string(output.Bytes())
				hello = strings.TrimSuffix(hello, "\n")
				hello = strings.ReplaceAll(hello, "\n", "\nlo: ")
				c.Exec(fmt.Sprintf(">&2 echo lo: %s", shellescape.Quote(hello)))
				if noEditTargets {
					c.Exec(fmt.Sprintf(">&2 echo lo: no editable targets, exit"))
					c.Close()
				}
				<-c.Closed()
			}(client)
		}
	}()

	hub := hub.New(db, bash)
	hub.LocalDomain = args.localdomain

	go http.Serve(net.Listener(bash), hub)
	go bash.Serve(l)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c

}

func hasRunning(addr string) (err error) {
	_, span := tracer.Start(context.Background(), "lcode-hub check")
	defer span.End()
	defer err2.Handle(&err)

	conn := To1(net.Dial("tcp", addr))
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(2 * time.Second))

	To1(io.WriteString(conn, "-1\n"))
	_sid := To1(io.ReadAll(conn))
	sid := string(_sid)

	if !strings.HasPrefix(sid, "lcode-hub") {
		err = fmt.Errorf("expect lcode-hub, but got %s", sid)
		return
	}

	span.SetStatus(codes.Error, "lcode-hub already has running")
	span.SetAttributes(attribute.String("version", sid))
	return
}
