package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/webdav"
)

func initWebdav(mux *http.ServeMux) {
	fs := webdav.Dir("/")
	lock := webdav.NewMemLS()

	const prefix = "/dav/"
	http.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {

		var hprefix = prefix

		fpath := strings.TrimPrefix(r.URL.Path, "/dav")
		deny := db.IsDeny(fpath)
		if deny {
			w.WriteHeader(403)
			fmt.Fprintf(w, "the folder is deny access.")
			return
		}
		// fix webdav dir list content include itself when prefix is not the real request path
		realURI := r.Header.Get("Httprelay-Proxy-Url")
		if realURI == "" {
			realURI = r.Header.Get("X-Forwarded-Uri")
		}
		if realURI != "" {
			uri, err := url.Parse(realURI)
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, "real uri parse failed")
				return
			}
			ss := strings.Split(uri.Path, prefix)
			if len(ss) < 2 {
				w.WriteHeader(400)
				fmt.Fprintf(w, "real uri parse failed.2")
				return
			}
			hprefix = ss[0] + prefix
			r.URL.Path = ss[0] + r.URL.Path
			fmt.Println(ss[0])
		}
		srv := &webdav.Handler{
			FileSystem: fs,
			LockSystem: lock,
			Prefix:     hprefix,
		}
		srv.ServeHTTP(w, r)
	})
}
