package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/net/nettest"
)

func makeUniqueID() (id string, err error) {
	rif, err := nettest.RoutedInterface("ip", net.FlagUp|net.FlagBroadcast)
	if err != nil {
		return
	}
	hostname, err := os.Hostname()
	if err != nil {
		return
	}
	mac := strings.Replace(rif.HardwareAddr.String(), ":", "_", 5)
	id = fmt.Sprintf("%s-%s", hostname, mac)
	return
}
