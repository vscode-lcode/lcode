package webdav

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/SierraSoftworks/multicast/v2"
	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
)

type Client struct {
	conn  net.Conn
	ID    string
	tasks *ttlcache.Cache[string, chan net.Conn]

	PWD    string
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
func (c *Client) Open(r *bufio.Reader) (err error) {
	defer err2.Handle(&err, func() {
		c.log(fmt.Errorf("bash client start failed: %w", err))
	})
	To(c.initServerAddr(r))
	To(c.initID(r))
	To(c.initPWD(r))
	go c.tasks.Start()
	go c.statsLocker.Start()
	return
}

func (c *Client) initServerAddr(r *bufio.Reader) (err error) {
	defer err2.Handle(&err, func() {
		c.log(fmt.Errorf("init server addr failed: %w", err))
	})
	io.WriteString(c.conn, "echo $2\n")
	line, _ := To2(r.ReadLine())
	c.ServerAddr = string(line)
	if c.ServerAddr == "" {
		laddr := c.conn.LocalAddr()
		addr := To1(net.ResolveTCPAddr(laddr.Network(), laddr.String()))
		c.ServerAddr = fmt.Sprintf("127.0.0.1/%d", addr.Port)
	}
	return
}

func (c *Client) initID(r *bufio.Reader) (err error) {
	defer err2.Handle(&err, func() {
		c.log(fmt.Errorf("got default init id failed: %w", err))
	})
	io.WriteString(c.conn, "echo $(2>/dev/null dd if=~/.lcode-id || echo 0)-$(2>/dev/null dd if=/proc/sys/kernel/hostname)\n")
	line, _ := To2(r.ReadLine())
	c.ID = string(line)
	return
}

func (c *Client) StoreID(id string) (err error) {
	cmd := fmt.Sprintf("echo -n %s > ~/.lcode-id\n", id)
	io.WriteString(c.conn, cmd)
	return
}

func (c *Client) initPWD(r *bufio.Reader) (err error) {
	defer err2.Handle(&err, func() {
		c.log(fmt.Errorf("init pwd failed: %w", err))
	})
	io.WriteString(c.conn, "cd $1;pwd\n")
	line, _ := To2(r.ReadLine())
	c.PWD = string(line)
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
