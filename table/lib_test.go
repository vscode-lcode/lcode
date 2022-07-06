package table_test

import (
	"os"
	"testing"
	"time"

	"github.com/shynome/lcode/table"
)

func TestTable(t *testing.T) {
	pwd, _ := os.Getwd()
	table := table.New(pwd)
	table.EmptyRulesFile()
	err := table.Open()
	if err != nil {
		t.Error(err)
		return
	}
	table.Allow("/a/b")
	time.Sleep(time.Second)
	if table.IsDeny("/a/b") {
		t.Error("should be allowed")
	}
	table.Deny("/a/b")
	time.Sleep(time.Second)
	if !table.IsDeny("/a/b") {
		t.Error("should be deny")
	}
	return
}
