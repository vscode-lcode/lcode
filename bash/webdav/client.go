package webdav

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/SierraSoftworks/multicast/v2"
	"github.com/alessio/shellescape"
	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
)

type Client struct {
	conn  net.Conn
	ID    string
	tasks *ttlcache.Cache[string, chan net.Conn]

	PWD    string // 也有可能是文件, 用途是用于限制文件访问
	Logger func(file *File, err error)
	closed *multicast.Channel[any]

	statsLocker *ttlcache.Cache[string, *StatWithLocker]

	ServerAddr string
}

func NewClient(conn net.Conn) *Client {
	c := &Client{
		conn: conn,
		tasks: ttlcache.New(
			ttlcache.WithTTL[string, chan net.Conn](10 * time.Second),
		),
		closed: multicast.New[any](),

		statsLocker: ttlcache.New(
			ttlcache.WithTTL[string, *StatWithLocker](2 * time.Second),
		),
	}
	c.tasks.OnEviction(func(ctx context.Context, er ttlcache.EvictionReason, i *ttlcache.Item[string, chan net.Conn]) {
		if i == nil {
			return
		}
		ch := i.Value()
		close(ch)
	})
	return c
}

func (c *Client) log(err error) {
	if c.Logger == nil {
		return
	}
	c.Logger(nil, err)
}
func (c *Client) Open(r *bufio.Reader, version string, id string) (err error) {
	defer err2.Handle(&err, func() {
		c.log(fmt.Errorf("bash client start failed: %w", err))
	})
	To(c.parseArgs(r, version))
	To(c.initServerAddr(r, id))
	To(c.initID(r))
	To(c.initPWD(r))
	go c.tasks.Start()
	go c.statsLocker.Start()
	return
}

func (c *Client) intFlag(version string) *flag.FlagSet {
	f := flag.NewFlagSet("lcode@"+version, flag.ContinueOnError)
	f.StringVar(&c.PWD, "pwd", ".", "工作目录")
	f.StringVar(&c.ServerAddr, "server", c.conn.LocalAddr().String(), "server addr")
	f.Bool("x", true, "仅用以分割bash参数, 不作其他用途")
	return f
}
func (c *Client) parseArgs(r *bufio.Reader, version string) (err error) {
	defer err2.Handle(&err, func() {
		c.log(fmt.Errorf("parse lcode args failed: %w", err))
	})
	To1(io.WriteString(c.conn, "echo $@\n"))
	line, _ := To2(r.ReadLine())

	f := c.intFlag(version)
	var output bytes.Buffer
	f.SetOutput(&output)

	if err = f.Parse(strings.Split(string(line), " ")); err != nil {
		err = ErrPrintHelp
		output := strings.ReplaceAll(string(output.Bytes()), "\n", "\nlo: ")
		cmd := fmt.Sprintf(">&2 echo lo: %s\n", shellescape.Quote(output))
		To1(io.WriteString(c.conn, cmd))
		return
	}

	if f.Arg(0) != "" {
		f.Set("pwd", f.Arg(0))
	}
	return
}

func (c *Client) initServerAddr(r *bufio.Reader, id string) (err error) {
	defer err2.Handle(&err, func() {
		if errors.Is(err, ErrNeedPrint) {
			return
		}
		err = ErrServerAddrIncorrect
	})
	addr, err := net.ResolveTCPAddr("tcp", c.ServerAddr)
	if err != nil {
		return ErrServerAddrParseFailed
	}
	c.ServerAddr = fmt.Sprintf("%s/%d", addr.IP.String(), addr.Port)

	tcp := fmt.Sprintf("4>&0 5>/dev/tcp/%s 3> >(>&5 dd bs=1 <&4) dd bs=1 <&5", c.ServerAddr)
	cmd := fmt.Sprintf("echo -1 | %s || echo /dev/null\n", tcp)
	To1(io.WriteString(c.conn, cmd))
	line, _ := To2(r.ReadLine())
	if string(line) != id {
		err = ErrServerAddrIncorrect
		return
	}
	return
}

func (c *Client) initID(r *bufio.Reader) (err error) {
	defer err2.Handle(&err, func() {
		c.log(fmt.Errorf("got default init id failed: %w", err))
	})
	cmd := "echo $(2>/dev/null dd if=~/.lcode-id || echo 0)-$(2>/dev/null dd if=/proc/sys/kernel/hostname)\n"
	To1(io.WriteString(c.conn, cmd))
	line, _ := To2(r.ReadLine())
	c.ID = string(line)
	return
}

func (c *Client) StoreID(id string) (err error) {
	cmd := fmt.Sprintf("echo -n %s > ~/.lcode-id\n", id)
	_, err = io.WriteString(c.conn, cmd)
	return
}

func (c *Client) initPWD(r *bufio.Reader) (err error) {
	defer err2.Handle(&err, func() {
		c.log(fmt.Errorf("init pwd failed: %w", err))
	})
	cmd := fmt.Sprintf("TZ=UTC0 ls -Ald --full-time %s || echo /dev/null | head -n 1\n", shellescape.Quote(c.PWD))
	To1(io.WriteString(c.conn, cmd))
	line, _ := To2(r.ReadLine())
	if string(line) == "/dev/null" {
		err = ErrEditTargetNotExists
		return
	}
	f := parseLsLine(line)
	if strings.HasPrefix(f.sys[7], "/") {
		c.PWD = f.sys[7]
		return
	}
	To1(io.WriteString(c.conn, "pwd\n"))
	pwd, _ := To2(r.ReadLine())
	c.PWD = filepath.Join(string(pwd), f.name)
	return
}

func (c *Client) Close() {
	c.tasks.DeleteAll()
	c.tasks.Stop()

	c.statsLocker.DeleteAll()
	c.statsLocker.Stop()

	c.conn.Close()
	c.closed.Close()
}

func (c *Client) Closed() <-chan any {
	return c.closed.Listen().C
}

var ErrClientIDRrequired = errors.New("client id is required")

func (c *Client) Exec(cmd string) (r net.Conn, err error) {
	if c.ID == "" {
		return nil, ErrClientIDRrequired
	}

	ch := make(chan net.Conn)
	tid := uuid.NewString()
	c.tasks.Set(tid, ch, ttlcache.NoTTL)

	c.exec(tid, cmd)

	conn, ok := <-ch
	if !ok {
		return nil, fmt.Errorf("timeout")
	}
	return conn, nil
}

func (c *Client) exec(tid string, cmd string) {
	ids := fmt.Sprintf("1%s:%s", tid, c.ID)
	f := strings.Join([]string{
		fmt.Sprintf("1>/dev/tcp/%s", c.ServerAddr),
		fmt.Sprintf("0>&1"),
		fmt.Sprintf("4> >(echo %s) 4>&1", ids),
	}, " ")
	cmd = fmt.Sprintf("%s 1> >(0>&1 %s) &\n", f, cmd)
	// fmt.Println("exec cmd", cmd)
	io.WriteString(c.conn, cmd)
}

func (c *Client) Recive(tid string, conn net.Conn) (err error) {
	task := c.tasks.Get(tid)
	if task == nil {
		return fmt.Errorf("task handle is gone")
	}
	task.Value() <- conn
	return
}

func (c *Client) ExecNoreply(cmd string) (err error) {
	defer err2.Handle(&err)
	eid := uuid.NewString()
	cmd = fmt.Sprintf("%s && echo -n %s", cmd, eid)
	r := To1(c.Exec(cmd))
	b := To1(io.ReadAll(r))

	if !strings.HasSuffix(string(b), eid) {
		err = fmt.Errorf("%s failed", cmd)
		return
	}
	return
}
