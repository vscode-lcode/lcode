package webdav

import (
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
	"golang.org/x/net/webdav"
)

type File struct {
	c      *Client
	name   string
	cursor int64

	locker *sync.RWMutex
	stat   fs.FileInfo

	readerInit    *sync.Once
	readerInitErr error
	reader        net.Conn

	writerInit    *sync.Once
	writerInitErr error
	writer        net.Conn

	no uint64
}

var _ webdav.File = (*File)(nil)

var no uint64 = 0

func OpenFile(c *Client, filename string) *File {
	return &File{
		no: atomic.AddUint64(&no, 1),

		c:    c,
		name: filename,

		locker:     &sync.RWMutex{},
		readerInit: &sync.Once{},
		writerInit: &sync.Once{},
	}
}

func (f *File) Close() error {
	if f.reader != nil {
		f.reader.Close()
	}
	if f.writer != nil {
		defer f.c.statsLocker.Delete(f.name)
		f.writer.Close()
		time.Sleep(200 * time.Millisecond) //等200ms, 让dd把末尾的输入内容写入到文件内
	}
	return nil
}

func (f *File) Read(p []byte) (n int, err error) {
	defer err2.Handle(&err, func() {
		if errors.Is(err, io.EOF) {
			return
		}
		f.log(fmt.Errorf("read err: %w", err))
	})
	stat := To1(f.Stat())
	if stat.IsDir() {
		return 0, io.EOF
	}
	f.readerInit.Do(func() {
		cmd := fmt.Sprintf("dd status=none if=%s skip=%d", shellescape.Quote(f.name), f.cursor)
		cmd = fmt.Sprintf("%s %s", cmd, "iflag=skip_bytes")
		f.reader, f.readerInitErr = f.c.Exec(cmd)
	})
	if f.readerInitErr != nil {
		err = f.readerInitErr
		return
	}
	n, err = f.reader.Read(p)
	f.cursor += int64(n)
	return
}

func (f *File) Write(p []byte) (n int, err error) {
	defer err2.Handle(&err, func() {
		fmt.Println("write err:", err)
	})
	f.writerInit.Do(func() {
		cmd := fmt.Sprintf("dd status=none of=%s seek=%d", shellescape.Quote(f.name), f.cursor)
		cmd = fmt.Sprintf("%s %s", cmd, "oflag=seek_bytes")
		f.writer, f.writerInitErr = f.c.Exec(cmd)
	})
	if f.writerInitErr != nil {
		err = f.writerInitErr
		return
	}
	n, err = f.writer.Write(p)
	f.cursor += int64(n)
	return
}

func (f *File) Seek(offset int64, whence int) (n int64, err error) {
	defer err2.Handle(&err, func() {
		f.log(fmt.Errorf("seek err: %w", err))
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
