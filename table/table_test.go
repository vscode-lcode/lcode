package table

import (
	"bytes"
	"io/ioutil"
	"testing"
)

const testTxt = `
+a
+b
-a
+a
-b
+c
`

func TestLoadTable(t *testing.T) {
	m := LoadTable(bytes.NewBufferString(testTxt))
	if _, ok := m["a"]; !ok {
		t.Error("a should be exist")
		return
	}
	if _, ok := m["b"]; ok {
		t.Error("b should be not exist")
		return
	}
	if _, ok := m["c"]; !ok {
		t.Error("c should be exist")
		return
	}
	t.Log(m)
}

func BenchmarkLoadTable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := LoadTable(bytes.NewBufferString(testTxt))
		if _, ok := m["a"]; !ok {
			b.Error("a should be exist")
			return
		}
		if _, ok := m["b"]; ok {
			b.Error("b should be not exist")
			return
		}
		if _, ok := m["c"]; !ok {
			b.Error("c should be exist")
			return
		}
	}
}

func TestLoadTableFromFile(t *testing.T) {
	b, err := ioutil.ReadFile("./allowed-dirs.table")
	if err != nil {
		t.Error(err)
		return
	}
	m := LoadTable(bytes.NewBuffer(b))
	t.Log(m)
}
