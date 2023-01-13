package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.T) {
	var running string
	f.StringVar(&running, "test.run", "", "golang test")
	f.Parse(os.Args[1:])
	if running != "^TestMain$" {
		return
	}
	// VERSION = "test"
	main()
}
