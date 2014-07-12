package main

import (
	"io"
	"kirisurf/ll/intercom"
)

func e2e_server_handler(socket io.ReadWriteCloser) {
	intercom.RunMultiplexServer(socket)
}
