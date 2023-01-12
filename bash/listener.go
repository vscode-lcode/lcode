package bash

import (
	"io"
	"net"
)

var _ net.Listener = (*Bash)(nil)

func (sh *Bash) toHandler(conn net.Conn) {
	sh.handlerLocker.Lock()
	defer sh.handlerLocker.Unlock()
	if len(sh.handler) == 0 {
		return
	}
	ch := sh.handler[0]
	sh.handler = sh.handler[1:]
	ch <- conn
}
func (sh *Bash) addHandler(ch chan net.Conn) {
	sh.handlerLocker.Lock()
	defer sh.handlerLocker.Unlock()
	sh.handler = append(sh.handler, ch)
}
func (sh *Bash) Close() error {
	sh.handlerLocker.Lock()
	defer sh.handlerLocker.Unlock()
	for _, ch := range sh.handler {
		close(ch)
	}
	return nil
}

func (sh *Bash) Accept() (net.Conn, error) {
	ch := make(chan net.Conn)
	sh.addHandler(ch)
	conn, ok := <-ch
	if !ok {
		return nil, net.ErrClosed
	}
	return conn, nil
}

func (sh *Bash) Addr() net.Addr {
	if sh.listener == nil {
		return nil
	}
	return sh.listener.Addr()
}

type BashConn struct {
	net.Conn
	Reader io.Reader
}

func NewBashConn(reader io.Reader, conn net.Conn) *BashConn {
	return &BashConn{
		Conn:   conn,
		Reader: reader,
	}
}

func (bc *BashConn) Read(b []byte) (n int, err error) {
	n, err = bc.Reader.Read(b)
	return
}
