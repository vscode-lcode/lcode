package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/shynome/httprelay-go"
	"github.com/shynome/lcode/table"
)

var LCODE_CONNECT string = "http://127.0.0.1:4349"

var args struct {
	Connect string
}

func init() {
	defer initDB()

	c := os.Getenv("LCODE_CONNECT")
	if c != "" {
		LCODE_CONNECT = c
	}

	flag.StringVar(&args.Connect, "c", LCODE_CONNECT, "the lcode hub connect addr")

}

var db *table.Table

func initDB() {
	var err error
	lcodeTmpdir := filepath.Join(os.TempDir(), "lcode-nuts")
	err = os.MkdirAll(lcodeTmpdir, 0777)
	if err != nil {
		panic(err)
	}
	db = table.New(lcodeTmpdir)
	err = db.Open()
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
