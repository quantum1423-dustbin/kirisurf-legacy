package main

import (
	"io"
	"kirisurf-legacy/ll/intercom"
)

func e2e_server_kinda_old_handler(socket io.ReadWriteCloser) {
	intercom.RunMultiplexSOCKSServer(socket)
}
