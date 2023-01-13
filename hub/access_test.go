package hub

import (
	"testing"

	"github.com/lainio/err2/assert"
	. "github.com/lainio/err2/try"
)

type testLC struct {
	id  string
	pwd string
}

func (lc testLC) RawID() string     { return lc.id }
func (lc testLC) PWD() string       { return lc.pwd }
func (lc testLC) Targets() []string { return []string{lc.pwd} }

func TestAllowDir(t *testing.T) {
	lc := testLC{"5-aaa", "/www/vvv/"}
	id := To1(hub.IDGenerator(lc))
	host := id.(*Client).ToHost()
	a1 := To1(hub.AllowDir(host, "/www/vvv/8888"))
	assert.NotNil(a1)
	a2 := To1(hub.AllowDir(host, "/www/8888"))
	assert.Equal(a2, nil)
	To(id.Close())
	a3 := To1(hub.AllowDir(host, "/www/vvv/8888"))
	assert.Equal(a3, nil)
}

func TestOrderBy(t *testing.T) {
	hosts := make([]Host, 0)
	host := Host{Namespace: "default", No: 0}
	To(hub.addHost(&host))
	defer hub.db.Delete(host)
	To(hub.db.Desc("no").Find(&hosts))
	t.Log(hosts)
}
