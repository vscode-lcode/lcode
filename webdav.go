package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/webdav"
)

func initWebdav(mux *http.ServeMux) {
	rootWebdav := &webdav.Handler{
		FileSystem: webdav.Dir("/"),
		LockSystem: webdav.NewMemLS(),
	}
	http.HandleFunc("/dav/", func(w http.ResponseWriter, r *http.Request) {
		// fix got dir but got file
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/dav")
		rewriteURLIfDir(r)
		if denyNotAllowedDir(w, r) {
			return
		}
		rootWebdav.ServeHTTP(w, r)
	})
}

func rewriteURLIfDir(r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/") {
		return
	}
	stat, err := os.Stat(r.URL.Path)
	if err != nil {
		return
	}
	if stat.IsDir() {
		r.URL.Path += "/"
		return
	}
}

func denyNotAllowedDir(w http.ResponseWriter, r *http.Request) (done bool) {
	deny := db.IsDeny(r.URL.Path)
	if deny {
		w.WriteHeader(403)
		fmt.Fprintf(w, "the folder is deny access.")
		return true
	}
	return
}
