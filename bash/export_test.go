package bash

import "testing"

func TestOnConnected(t *testing.T) {
	c := StartTestClient(l.Addr())
	client := <-bash.Connected()
	t.Log(client)
	client.Close()
	_ = bash
	err := c.Wait()
	t.Log(err)
}
