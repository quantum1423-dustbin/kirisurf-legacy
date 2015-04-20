package main

import (
	"io"
	"kirisurf-legacy/ll/intercom"
)

func e2e_server_handler(socket io.ReadWriteCloser) {
	intercom.RunMultiplexServer(socket)
}
