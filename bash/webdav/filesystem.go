package webdav

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/alessio/shellescape"
	"github.com/jellydator/ttlcache/v3"
	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/util/err0"
	"go.opentelemetry.io/otel"
	"golang.org/x/net/webdav"
)

var _ webdav.FileSystem = (*Client)(nil)

func (c *Client) Mkdir(ctx context.Context, name string, perm os.FileMode) (err error) {
	_, span := otel.Tracer(name).Start(c.Ctx, "fs mkdir")
	defer span.End()
	defer err0.Record(&err, span)

	cmd := fmt.Sprintf("mkdir -p %s", shellescape.Quote(name))
	To(c.ExecNoreply(cmd))
	return
}
func (c *Client) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (f webdav.File, err error) {
	_, span := otel.Tracer(name).Start(c.Ctx, "fs openfile")
	defer span.End()
	defer err0.Record(&err, span)

	f = OpenFile(c, name)
	return
}
func (c *Client) RemoveAll(ctx context.Context, name string) (err error) {
	_, span := otel.Tracer(name).Start(c.Ctx, "fs remove all")
	defer span.End()
	defer err0.Record(&err, span)

	defer c.statsLocker.DeleteAll()

	cmd := fmt.Sprintf("rm -rf %s", shellescape.Quote(name))
	To(c.ExecNoreply(cmd))
	return
}
func (c *Client) Rename(ctx context.Context, oldName, newName string) (err error) {
	_, span := otel.Tracer(name).Start(c.Ctx, "fs rename")
	defer span.End()
	defer err0.Record(&err, span)

	defer c.statsLocker.Delete(oldName)
	defer c.statsLocker.Delete(newName)

	cmd := fmt.Sprintf("mv %s %s", shellescape.Quote(oldName), shellescape.Quote(newName))
	To(c.ExecNoreply(cmd))
	return
}

func (c *Client) Stat(ctx context.Context, name string) (f os.FileInfo, err error) {
	_, span := otel.Tracer(name).Start(c.Ctx, "fs stat")
	defer span.End()
	defer err0.Record(&err, span)

	sl := c.StatLocker(name)
	sl.locker.RLock()
	f = sl.stat
	sl.locker.RUnlock()
	if f != nil {
		return
	}

	sl.locker.Lock()
	defer sl.locker.Unlock()
	f, err = OpenFile(c, name)._Stat()
	if err != nil {
		return
	}
	sl.stat = f

	return
}

type StatWithLocker struct {
	locker *sync.RWMutex
	stat   fs.FileInfo
}

func (client *Client) StatLocker(name string) (sl *StatWithLocker) {
	item := client.statsLocker.Get(name)
	if item == nil {
		sl = &StatWithLocker{locker: &sync.RWMutex{}, stat: nil}
		client.statsLocker.Set(name, sl, ttlcache.DefaultTTL)
	} else {
		sl = item.Value()
	}
	return
}
