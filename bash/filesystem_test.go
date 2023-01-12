package bash

import (
	"context"
	"os"
	"testing"

	"github.com/lainio/err2/assert"
	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/bash/webdav"
)

func TestFileReaddir(t *testing.T) {
	item := bash.clients.Get(TestClientID("0").String())
	assert.NotNil(item)
	client := item.Value()
	f := webdav.OpenFile(client, ".")
	files := To1(f.Readdir(-1))
	t.Log(files)
	return
}

func TestFileinfo(t *testing.T) {
	f := To1(os.Stat("/tmp/a"))
	m := f.Mode()
	t.Log(m)
	f2 := To1(client.Stat(context.Background(), "/tmp/a"))
	a, b := fileinfo2map(f), fileinfo2map(f2)
	assert.Equal(b, a)
	return
}
