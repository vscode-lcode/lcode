package webdav

import (
	"io/fs"
	"os"
	"strings"
)

var modeMap = map[string]fs.FileMode{
	"r": 0b100,
	"w": 0b010,
	"x": 0b001,
	"d": os.ModeDir,
}

func mode2int(mode string) fs.FileMode {
	m, ok := modeMap[mode]
	if !ok {
		return 0
	}
	return m
}

func parseFilemode(filemode string) (mode fs.FileMode) {
	modes := strings.Split(filemode, "")
	mode = 0
	for i, m := range modes[1:] {
		n := mode2int(m)
		n = n << ((2 - int(i/3)) * 3)
		mode = mode | n
	}
	mode = mode | mode2int(modes[0])
	return
}
