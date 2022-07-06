package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func genVscodeLink(id string, w string) string {
	w = strings.TrimPrefix(w, "/")
	link := fmt.Sprintf("vscode://lcode.hub/%s/%s", id, w)
	return link
}

func reqOpen(link string) (err error) {
	reqLink, err := url.Parse(LCODE_CONNECT)
	if err != nil {
		return
	}

	reqLink.Path += "open"

	q := reqLink.Query()
	q.Set("link", link)

	reqLink.RawQuery = q.Encode()

	reqLinkStr := reqLink.String()
	_, err = http.Get(reqLinkStr)
	return
}
