package hub

import (
	"fmt"
	"net"
	"net/http"
	"testing"

	. "github.com/lainio/err2/try"
)

func TestWebdav(t *testing.T) {
	var l net.Listener = hub.bash
	fmt.Println("webdav sever: ", l.Addr())
	To(http.Serve(l, hub))
}
