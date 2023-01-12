package bash

import "github.com/vscode-lcode/lcode/v2/bash/webdav"

func (sh *Bash) Get(id string) *webdav.Client {
	item := sh.clients.Get(id)
	if item == nil {
		return nil
	}
	return item.Value()
}

func (sh *Bash) Connected() <-chan *webdav.Client {
	return sh.connected.Listen().C
}
