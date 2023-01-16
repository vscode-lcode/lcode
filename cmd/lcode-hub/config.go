package main

import (
	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
	"xorm.io/xorm"
)

type Config struct {
	Id    int64
	Name  string `xorm:"notnull unique"`
	Value string
}

func getConfig(db *xorm.Engine, name string, defaultValueGen func() string) (value string, err error) {
	defer err2.Handle(&err)
	var c Config = Config{Name: name}
	session := db.NewSession()
	defer session.Close()
	has := To1(db.Get(&c))
	if !has {
		c.Value = defaultValueGen()
		To1(db.InsertOne(&c))
	}
	To(session.Commit())
	value = c.Value
	return
}
