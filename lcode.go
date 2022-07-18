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
	flag.Parse()

	defer db.Dispose()

	proxy := httprelay.NewProxy(args.Connect)
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

	p := proxy.Auth.ID + path.Join("/dav/", codedir)
	vscodeLink, err := getOpenLink(p)
	if err != nil {
		fmt.Printf("get vscode open link failed. err: %e\n", err)
		return
	}

	stat, err := os.Stat(codedir)
	if err != nil {
		fmt.Println("can't get file stat")
		return
	}
	if !stat.IsDir() {
		vscodeLink += "#file"
	}
	fmt.Println(vscodeLink)
	// go reqOpen(vscodeLink)
	go proxy.Serve(nil)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
