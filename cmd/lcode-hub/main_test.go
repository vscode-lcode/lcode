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

func TestProgram(t *testing.T) {
	var running string
	f.StringVar(&running, "test.run", "", "golang test")
	f.Parse(os.Args[1:])
	if running != "^TestProgram$" {
		return
	}
	main()
}
