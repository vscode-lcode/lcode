package bash

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/lainio/err2/try"
	. "github.com/lainio/err2/try"
	"golang.org/x/net/webdav"
)

func TestWebdav(t *testing.T) {
	var l net.Listener = bash
	h := &webdav.Handler{
		FileSystem: client,
		LockSystem: webdav.NewMemLS(),
	}
	fmt.Println(l.Addr().String())
	To(http.Serve(l, h))
	return
}

func TestWebdavStat(t *testing.T) {
	f := "/opt/v2ray/v2ctl"
	fs := webdav.Dir("/")
	fname := fileinfo2map(To1(fs.Stat(context.Background(), f)))
	fname2 := fileinfo2map(To1(client.Stat(context.Background(), f)))
	t.Log(fname, fname2)
}

func BenchmarkStat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = client.Stat(context.Background(), fmt.Sprintf("/tmp/noooo-%d", i))
	}
}

func TestWebdavReaddir(t *testing.T) {
	f := "/home/shynome/aa"
	fs := webdav.Dir("/")
	dirs1, err := To1(fs.OpenFile(context.Background(), f, 0, 0)).Readdir(0)
	To(err)
	dirs2, err := To1(client.OpenFile(context.Background(), f, 0, 0)).Readdir(0)
	To(err)
	t.Log(dirs1, dirs2)
}

func fileinfo2map(f os.FileInfo) string {
	v := map[string]any{
		"name":     f.Name(),
		"size":     f.Size(),
		"mode":     f.Mode(),
		"mod-time": f.ModTime().UnixNano(),
		"dir":      f.IsDir(),
	}
	b := try.To1(json.Marshal(v))
	return string(b)
}
