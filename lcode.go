package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"

	"github.com/shynome/httprelay-go"
)

func main() {
	defer db.Dispose()

	proxy := httprelay.NewProxy(LCODE_CONNECT)
	initProxy(proxy)
	initWebdav(http.DefaultServeMux)

	var codedir = "."
	flag.Parse()
	var args = flag.Args()
	if len(args) >= 1 {
		codedir = args[0]
	}
	if strings.HasPrefix(codedir, "~") {
		homedir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		codedir = strings.Replace(codedir, "~", homedir, 1)
	}
	var err error
	codedir, err = filepath.Abs(codedir)
	if err != nil {
		panic(err)
	}

	db.Allow(codedir)
	defer db.Deny(codedir)

	vscodeLink := genVscodeLink(proxy.Auth.ID, path.Join("/dav/", codedir))

	fmt.Println(vscodeLink)
	// go reqOpen(vscodeLink)
	go proxy.Serve(nil)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
