package hub

import (
	"fmt"
	"math"
	"testing"

	"github.com/lainio/err2/assert"
	. "github.com/lainio/err2/try"
)

func TestParseIDRaw(t *testing.T) {
	var raw = fmt.Sprintf("%d-55555", math.MaxUint32)
	host := parseIDRaw(raw)
	assert.Equal(host.No, math.MaxUint32)
	assert.Equal(host.Namespace, "55555")
}

func TestAddHost(t *testing.T) {
	id := To1(hub.IDGenerator(testLC{"1-shy-matx", "/"}))
	defer id.Close()
	t.Log(id.No())
	id2 := To1(hub.IDGenerator(testLC{"1-shy-matx", "/"}))
	defer id2.Close()
	t.Log("pass")
}
