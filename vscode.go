package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type uQuery map[string]string

func makeReqLink(upath string, query uQuery) (link string, err error) {
	r, err := url.Parse(args.Connect)
	if err != nil {
		return
	}
	r.Path = path.Join(r.Path, upath)
	q := r.Query()
	for k := range query {
		q.Set(k, query[k])
	}
	r.RawQuery = q.Encode()
	link = r.String()
	return
}

func genVscodeLink(id string, w string) string {
	w = strings.TrimPrefix(w, "/")
	link := fmt.Sprintf("vscode://lcode.hub/%s/%s", id, w)
	return link
}

func getOpenLink(path string) (link string, err error) {
	rlink, err := makeReqLink("/open-link", uQuery{"path": path})
	if err != nil {
		return
	}
	resp, err := http.Get(rlink)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusAccepted {
		err = fmt.Errorf("get open link failed from hub. resp status: %s", resp.Status)
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	link = string(b)
	return
}

func reqOpen(link string) (err error) {
	rlink, err := makeReqLink("/open", uQuery{"link": link})
	if err != nil {
		return
	}

	_, err = http.Get(rlink)
	return
}
