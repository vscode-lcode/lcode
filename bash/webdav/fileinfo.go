package webdav

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type FileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
	sys     []string
}

var _ fs.FileInfo = NewFileInfo([]string{})

var lsLineRegex = regexp.MustCompile(`^([drwx-]{10}) +(\d+) +(\S+ +\S+) +(\d+) (\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(\.\d{9}|) \+\d{4}) +(.+)$`)

func parseLsLine(line []byte) FileInfo {
	info := lsLineRegex.FindStringSubmatch(string(line))
	return NewFileInfo(info)
}

func NewFileInfo(f []string) (info FileInfo) {
	info = FileInfo{sys: f}
	if info.IsNil() {
		return
	}
	info.name = filepath.Base(f[7])
	info.size, _ = strconv.ParseInt(f[4], 10, 64)
	info.mode = parseFilemode(f[1])
	info.modTime = parseModtime(f[5])
	return info
}

var timeRegex = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2})(\.\d{9}|) \+\d{4}$`)

func parseModtime(ts string) (t time.Time) {
	arr := timeRegex.FindStringSubmatch(ts)
	if len(arr) != 3 {
		return
	}
	tf := fmt.Sprintf("%sT%sZ", arr[1], arr[2])
	t, err := time.Parse(time.RFC3339Nano, tf)
	if err != nil {
		return
	}
	return t
}

func (f FileInfo) IsNil() bool { return len(f.sys) != 8 }

func (f FileInfo) Name() string       { return f.name }
func (f FileInfo) Size() int64        { return f.size }
func (f FileInfo) Mode() fs.FileMode  { return f.mode }
func (f FileInfo) ModTime() time.Time { return f.modTime }
func (f FileInfo) IsDir() bool        { return f.mode.IsDir() }
func (f FileInfo) Sys() any           { return f.sys }
