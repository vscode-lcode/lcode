package table

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type AllowedDirs map[string]bool

type Table struct {
	lock        *sync.RWMutex
	allowedDirs AllowedDirs
	rulesFile   string
	stopWatch   chan bool
}

func New(dir string) *Table {
	table := &Table{
		lock:        &sync.RWMutex{},
		allowedDirs: AllowedDirs{},
		stopWatch:   make(chan bool),
		rulesFile:   filepath.Join(dir, "/allowed-dirs.table"),
	}
	return table
}

func (t *Table) Open() (err error) {
	err = t.LoadTable()
	go t.Watch()
	return
}

func (t *Table) Allow(dir string) {
	t.writeToFile("+", dir)

	t.lock.Lock()
	defer t.lock.Unlock()
	t.allowedDirs[dir] = true
}

func (t *Table) Deny(dir string) {
	t.writeToFile("-", dir)

	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.allowedDirs, dir)
}

func (t *Table) writeToFile(flag, dir string) {
	f, err := os.OpenFile(t.rulesFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	rule := flag + dir + "\n"
	if _, err := f.WriteString(rule); err != nil {
		panic(err)
	}
}
func (t *Table) EmptyRulesFile() error {
	return os.Remove(t.rulesFile)
}

func (t *Table) IsDeny(dir string) (deny bool) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	deny = true
	for d := range t.allowedDirs {
		if strings.HasPrefix(dir, d) {
			return false
		}
	}
	return true
}

func (t *Table) GetAllowedDirs() map[string]bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.allowedDirs
}

func (t *Table) Dispose() (err error) {
	t.stopWatch <- true

	t.LoadTable()

	var count int = 0
	for range t.allowedDirs {
		count++
	}
	if count == 0 {
		if err = t.EmptyRulesFile(); err != nil {
			return
		}
	}
	return
}

func (t *Table) Watch() {
	ticker := time.NewTicker(time.Millisecond * 500)
	lastMtime := t.getRulesFileStat().ModTime()
	for {
		select {
		case <-t.stopWatch:
			ticker.Stop()
			return
		case <-ticker.C:
			stat := t.getRulesFileStat()
			mtime := stat.ModTime()
			if lastMtime.Equal(mtime) {
				continue
			}
			lastMtime = mtime
			if err := t.LoadTable(); err != nil {
				panic(err)
			}
		}
	}
}
func (t *Table) getRulesFileStat() fs.FileInfo {
	stat, err := os.Stat(t.rulesFile)
	if err != nil {
		panic(err)
	}
	return stat
}

func (t *Table) LoadTable() (err error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	f, err := os.OpenFile(t.rulesFile, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	m := LoadTable(f)

	t.allowedDirs = m
	return
}

func LoadTable(r io.Reader) AllowedDirs {
	m := AllowedDirs{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "+") {
			dir := strings.TrimPrefix(line, "+")
			m[dir] = true
		}
		if strings.HasPrefix(line, "-") {
			dir := strings.TrimPrefix(line, "-")
			_, ok := m[dir]
			if ok {
				delete(m, dir)
			}
		}
	}
	return m
}
