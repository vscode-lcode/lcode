package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/shynome/httprelay-go"
	"github.com/shynome/lcode/table"
)

var LCODE_CONNECT string = "http://127.0.0.1:4349"

var db *table.Table

func init() {
	c := os.Getenv("LCODE_CONNECT")
	if c != "" {
		LCODE_CONNECT = c
	}

	lcodeTmpdir := filepath.Join(os.TempDir(), "lcode-nuts")
	db = table.New(lcodeTmpdir)
	err := db.Open()
	if err != nil {
		panic(err)
	}
}

func initProxy(proxy *httprelay.Proxy) {
	id, err := makeUniqueID()
	if err != nil {
		log.Fatal(err)
	}
	proxy.Auth.ID = id
	proxy.Auth.Secret = id
	if proxy.Parallel < 4 {
		proxy.Parallel = 4
	}
	proxy.Println = func(i ...interface{}) {}
}
