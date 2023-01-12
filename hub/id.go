package hub

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/jellydator/ttlcache/v3"
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/bash"
	"golang.org/x/net/webdav"
	"xorm.io/builder"
)

var _ bash.IDGenerator = (*Hub)(nil).IDGenerator

func (hub *Hub) IDGenerator(idRaw string, pwd string) (id bash.ID, err error) {
	defer err2.Handle(&err)
	host := parseIDRaw(idRaw)
	if host.No == 0 {
		To(hub.addHost(&host))
	}
	client := &Client{
		Namespace: host.Namespace,
		No2:       host.No,
		Workdir:   pwd,
	}
	To(hub.addClient(client))
	hub.setHostLocker(client.ToHost())
	To(hub.UpdateAllowedDir(*client))
	client.hub = hub
	id = client
	return
}

func (hub *Hub) setHostLocker(host Host) {
	item := hub.lockers.Get(host.String())
	if item != nil {
		return
	}
	locker := webdav.NewMemLS()
	hub.lockers.Set(host.String(), locker, ttlcache.NoTTL)
	return
}
func (hub *Hub) addHost(host *Host) (err error) {
	defer err2.Handle(&err)
	session := hub.db.NewSession()
	defer session.Close()
	To(session.Begin())
	total := To1(session.Where(builder.Eq{"namespace": host.Namespace}).Count(new(Host)))
	host.No = uint32(total) + 1
	To1(session.Insert(host))
	To(session.Commit())
	return
}

func (hub *Hub) addClient(client *Client) (err error) {
	_, err = hub.clientDB.Insert(client)
	return
}

// 05671-anystring
var idRawRegexp = regexp.MustCompile(`^(\d+)-(.+)$`)

func parseIDRaw(raw string) (host Host) {
	arr := idRawRegexp.FindStringSubmatch(raw)
	switch l := len(arr); l {
	case 2:
		arr = append(arr, "default")
		fallthrough
	case 3:
		no := To1(strconv.ParseUint(arr[1], 10, 32))
		host.No = uint32(no)
		host.Namespace = arr[2]
	default:
		host = Host{Namespace: "default", No: 0}
	}
	return
}

var _ bash.ID = (*Client)(nil)

func (id Client) NameSapce() string { return id.Namespace }
func (id Client) No() string        { return fmt.Sprint(id.No2) }
func (id Client) String() string {
	return fmt.Sprintf("%d-%s-%s", id.Id, id.NameSapce(), id.No())
}
func (id Client) Close() (err error) {
	defer err2.Handle(&err)
	if id.hub == nil {
		return
	}
	To1(id.hub.clientDB.ID(id.Id).Delete(new(Client)))
	To(id.hub.UpdateAllowedDir(id))
	return
}

func (id Client) ToHost() Host {
	return Host{
		Namespace: id.Namespace,
		No:        id.No2,
	}
}
