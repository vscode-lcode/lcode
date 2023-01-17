package webdav

import (
	"context"
	"fmt"
	"os"

	"github.com/alessio/shellescape"
	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/util/err0"
	"golang.org/x/net/webdav"
)

var _ webdav.FileSystem = (*Client)(nil)

func (c *Client) Mkdir(ctx context.Context, name string, perm os.FileMode) (err error) {
	_, span := tracer.Start(c.Ctx, "fs mkdir")
	defer span.End()
	defer err0.Record(&err, span)

	cmd := fmt.Sprintf("mkdir -p %s", shellescape.Quote(name))
	To(c.ExecNoreply(cmd))
	return
}
func (c *Client) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (f webdav.File, err error) {
	_, span := tracer.Start(c.Ctx, "fs openfile")
	defer span.End()
	defer err0.Record(&err, span)

	f = OpenFile(c, name)
	return
}
func (c *Client) RemoveAll(ctx context.Context, name string) (err error) {
	_, span := tracer.Start(c.Ctx, "fs remove all")
	defer span.End()
	defer err0.Record(&err, span)

	defer c.statsCache.DeleteAll()

	cmd := fmt.Sprintf("rm -rf %s", shellescape.Quote(name))
	To(c.ExecNoreply(cmd))
	return
}
func (c *Client) Rename(ctx context.Context, oldName, newName string) (err error) {
	_, span := tracer.Start(c.Ctx, "fs rename")
	defer span.End()
	defer err0.Record(&err, span)

	defer c.statsCache.Delete(oldName)
	defer c.statsCache.Delete(newName)

	cmd := fmt.Sprintf("mv %s %s", shellescape.Quote(oldName), shellescape.Quote(newName))
	To(c.ExecNoreply(cmd))
	return
}

func (c *Client) Stat(ctx context.Context, name string) (f os.FileInfo, err error) {
	_, span := tracer.Start(c.Ctx, "fs stat")
	defer span.End()
	defer err0.Record(&err, span)

	item := c.statsCache.Get(name)
	if item != nil {
		f = item.Value()
		return
	}

	f, err = OpenFile(c, name)._Stat()
	if err != nil {
		return
	}
	c.statsCache.Set(name, f, 0)

	return
}
