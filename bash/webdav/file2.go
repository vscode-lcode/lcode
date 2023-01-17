package webdav

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alessio/shellescape"
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/util/err0"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/webdav"
)

type File struct {
	Ctx context.Context

	c      *Client
	name   string
	cursor int64

	statusInit *sync.Once
	stat       fs.FileInfo

	readerInit *sync.Once
	reader     net.Conn

	writerInit *sync.Once
	writer     net.Conn

	no uint64
}

var _ webdav.File = (*File)(nil)

var no uint64 = 0

func OpenFile(c *Client, filename string) *File {
	ctx, span := tracer.Start(c.Ctx, "file open")
	span.SetAttributes(
		attribute.String("filepath", filename),
	)
	span.AddEvent("nolog")
	return &File{
		Ctx: ctx,

		no: atomic.AddUint64(&no, 1),

		c:    c,
		name: filename,

		statusInit: &sync.Once{},
		readerInit: &sync.Once{},
		writerInit: &sync.Once{},
	}
}

func (f *File) Close() error {
	defer trace.SpanFromContext(f.Ctx).End()

	if f.reader != nil {
		f.reader.Close()
	}
	if f.writer != nil {
		defer f.c.statsCache.Delete(f.name)
		f.writer.Close()
		time.Sleep(200 * time.Millisecond) //等200ms, 让dd把末尾的输入内容写入到文件内
	}
	return nil
}

func (f *File) Read(p []byte) (n int, err error) {
	_, span := tracer.Start(f.Ctx, "file read")
	defer span.End()
	defer err2.Handle(&err, func() {
		if errors.Is(err, io.EOF) {
			return
		}
		err0.Record(&err, span)
	})

	stat := To1(f.Stat())
	if stat.IsDir() {
		return 0, io.EOF
	}
	f.readerInit.Do(func() {
		cmd := fmt.Sprintf("dd if=%s skip=%d", shellescape.Quote(f.name), f.cursor)
		cmd = fmt.Sprintf("%s %s", cmd, "iflag=skip_bytes")
		f.reader, err = f.c.Exec(cmd)
	})
	if err != nil {
		f.readerInit = &sync.Once{}
		return
	}
	f.reader.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer f.reader.SetReadDeadline(time.Time{})
	n, err = f.reader.Read(p)
	f.cursor += int64(n)
	return
}

func (f *File) Write(p []byte) (n int, err error) {
	_, span := tracer.Start(f.Ctx, "file write")
	defer span.End()
	defer err0.Record(&err, span)

	f.writerInit.Do(func() {
		cmd := fmt.Sprintf("dd of=%s seek=%d", shellescape.Quote(f.name), f.cursor)
		cmd = fmt.Sprintf("%s %s", cmd, "oflag=seek_bytes")
		f.writer, err = f.c.Exec(cmd)
	})
	if err != nil {
		f.writerInit = &sync.Once{}
		return
	}
	f.writer.SetWriteDeadline(time.Now().Add(5 * time.Second))
	defer f.writer.SetWriteDeadline(time.Time{})
	n, err = f.writer.Write(p)
	f.cursor += int64(n)
	return
}

func (f *File) Seek(offset int64, whence int) (n int64, err error) {
	_, span := tracer.Start(f.Ctx, "file seek")
	defer span.End()
	defer err2.Handle(&err, func() {
		span.RecordError(err)
	})
	switch whence {
	case io.SeekStart:
		f.cursor = offset
	case io.SeekCurrent:
		f.cursor += offset
	case io.SeekEnd:
		stat := To1(f.Stat())
		f.cursor = stat.Size() + offset
	}
	n = f.cursor
	f.readerInit = &sync.Once{}
	f.writerInit = &sync.Once{}
	return
}
