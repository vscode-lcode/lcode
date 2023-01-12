package hub

import (
	"strings"

	"github.com/jellydator/ttlcache/v3"
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
	"xorm.io/builder"
)

func (hub *Hub) AllowDir(host Host, filepath string) (client *Client, err error) {
	item := hub.allowedDirs.Get(host.String())
	if item == nil {
		return
	}
	dirs := item.Value()
	for _, item := range dirs {
		if allow := strings.HasPrefix(filepath, item.Workdir); allow {
			client = &item
			return
		}
	}
	return
}

func (hub *Hub) UpdateAllowedDir(client Client) (err error) {
	defer err2.Handle(&err)
	dirs := make([]Client, 0)
	host := client.ToHost()
	q := builder.Eq{"no": host.No, "namespace": host.Namespace}
	To(hub.clientDB.Where(q).Desc("workdir").Find(&dirs))
	hub.allowedDirs.Set(host.String(), dirs, ttlcache.NoTTL)
	return
}
