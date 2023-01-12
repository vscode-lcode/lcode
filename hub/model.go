package hub

import (
	"fmt"

	"xorm.io/xorm"
)

type Host struct {
	Id        int64
	Namespace string `xorm:"notnull unique(host-id) index"`
	No        uint32 `xorm:"notnull unique(host-id)"`
}

func (h Host) String() string { return fmt.Sprintf("%d-%s", h.No, h.Namespace) }

type Client struct {
	hub *Hub

	Id        int64
	Namespace string `xorm:"notnull"`
	No2       uint32 `xorm:"notnull 'no'"`
	Workdir   string `xorm:"notnull"`
}

func Sync(eg *xorm.Engine) (err error) {
	return eg.Sync(new(Host))
}
