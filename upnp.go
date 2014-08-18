package main

import (
	"fmt"
	"os/exec"
	"strings"
)

var upnpc_loc = fmt.Sprintf("%s.%s", "./utilities/upnpc", EXESUF)

func UPnPForwardAddr(addr string) error {
	port := strings.Split(addr, ":")[1]
	cmd := exec.Command(upnpc_loc, "-r", port, "tcp")
	_, err := cmd.CombinedOutput()
	return err
}
