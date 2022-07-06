package main

import (
	"testing"
)

func TestGenVscodeLink(t *testing.T) {
	var expectLink = "vscode://lcode.hub/aaaa/home/xwww"
	var gotLink = genVscodeLink("aaaa", "/home/xwww")
	if gotLink != expectLink {
		t.Errorf("expect: %s, got: %s", expectLink, gotLink)
		return
	}
}

func TestReqOpen(t *testing.T) {
	var vscodeLink = "vscode://lcode.hub/aaaa/home/xwww"
	reqOpen(vscodeLink)
}
