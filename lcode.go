package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"

	"github.com/shynome/httprelay-go"
)

func main() {
	defer db.Dispose()

	proxy := httprelay.NewProxy(LCODE_CONNECT)
	initProxy(proxy)
	initWebdav(http.DefaultServeMux)

	pwd, _ := os.Getwd()
	var codedir = "."
	if len(os.Args) >= 2 {
		codedir = os.Args[1]
	}
	codedir = filepath.Join(pwd, codedir)

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
