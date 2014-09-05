package main

import (
	"io"
	"kirisurf/ll/intercom"
)

func e2e_server_kinda_old_handler(socket io.ReadWriteCloser) {
	intercom.RunMultiplexSOCKSServer(socket)
}
