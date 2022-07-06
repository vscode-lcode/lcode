package main

import "testing"

func TestMakeUniqueID(t *testing.T) {
	uqid, err := makeUniqueID()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(uqid)
}
