package webdav

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alessio/shellescape"
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/util/err0"
)

func (f *File) Readdir(count int) (files []fs.FileInfo, err error) {
	_, span := tracer.Start(f.Ctx, "file readdir")
	defer span.End()
	defer err0.Record(&err, span)

	files = To1(f.readdir(count))
	for _, stat := range files {
		fname := filepath.Join(f.name, stat.Name())
		f.c.statsCache.Set(fname, stat, 0)
	}
	return
}

func (f *File) readdir(n int) (files []fs.FileInfo, err error) {
	defer err2.Handle(&err)
	cmd := fmt.Sprintf("TZ=UTC0 ls -Al --full-time %s", shellescape.Quote(f.name))
	conn := To1(f.c.Exec(cmd))
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	r := bufio.NewReader(conn)
	for {
		line, _, err := r.ReadLine()
		if IsEOF(err) {
			break
		}
		fileinfo := parseLsLine(line)
		if !fileinfo.IsNil() {
			files = append(files, fileinfo)
		}
		if n > 0 && len(files) >= n {
			break
		}
	}
	return
}

func (f *File) Stat() (finfo fs.FileInfo, err error) {
	return f.c.Stat(context.Background(), f.name)
}
func (f *File) _Stat() (finfo fs.FileInfo, err error) {
	f.statusInit.Do(func() {
		f.stat, err = f._GetStat()
	})
	if err != nil {
		f.statusInit = &sync.Once{}
		return
	}
	finfo = f.stat
	return
}
func (f *File) _GetStat() (finfo fs.FileInfo, err error) {
	_, span := tracer.Start(f.Ctx, "file stat")
	defer span.End()
	defer err2.Handle(&err, func() {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		err0.Record(&err, span)
	})

	cmd := fmt.Sprintf("TZ=UTC0 ls -Ald --full-time %s", shellescape.Quote(f.name))
	conn := To1(f.c.Exec(cmd))
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	r := bufio.NewReader(conn)
	line, _, err := r.ReadLine()
	if IsEOF(err) {
		err = os.ErrNotExist
		return
	}
	rfinfo := parseLsLine(line)
	if rfinfo.IsNil() {
		err = fmt.Errorf("get file %s stats failed", f.name)
		return
	}
	finfo = rfinfo
	return
}
