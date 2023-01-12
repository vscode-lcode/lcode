package hub

import (
	"github.com/jellydator/ttlcache/v3"
	"github.com/vscode-lcode/lcode/v2/bash"
	"golang.org/x/net/webdav"
	"xorm.io/xorm"
)

type Hub struct {
	db          *xorm.Engine
	bash        *bash.Bash
	lockers     *ttlcache.Cache[string, webdav.LockSystem]
	allowedDirs *ttlcache.Cache[string, []Client]
	LocalDomain string
}

func New(db *xorm.Engine, bash *bash.Bash) *Hub {
	hub := &Hub{
		db:   db,
		bash: bash,
		lockers: ttlcache.New(
			ttlcache.WithTTL[string, webdav.LockSystem](ttlcache.NoTTL),
		),
		allowedDirs: ttlcache.New(
			ttlcache.WithTTL[string, []Client](ttlcache.NoTTL),
		),
		LocalDomain: ".lo.shynome.com",
	}
	bash.IDGenerator = hub.IDGenerator
	return hub
}
