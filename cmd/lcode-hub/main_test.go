package main

import (
	"os"
	"testing"
)

var changeDefaultArgs int = func() int {
	// VERSION = "test"
	defaultLogLv = "11"
	return 1
}()

func TestMain(m *testing.T) {
	var running string
	f.StringVar(&running, "test.run", "", "golang test")
	f.Parse(os.Args[1:])
	if running != "^TestMain$" {
		return
	}
	main()
}
