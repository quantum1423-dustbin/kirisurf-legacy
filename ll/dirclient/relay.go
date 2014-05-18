// relay.go
package dirclient

import (
	"fmt"
	"net/http"
	"time"
)

const (
	PROTVER = 300
)

var ISRELAY = false

// This function blocks. Run in goroutine.
func RunRelay(port int, keyhash string, isexit bool) {
	ISRELAY = true
	ieflag := 0
	if isexit {
		ieflag = 1
	}
	RefreshDirectory()
	url := fmt.Sprintf("%s/upload?port=%d&protocol=%d&keyhash=%s&exit=%d",
		DIRADDR,
		port, PROTVER, keyhash, ieflag)
	r, e := http.Get(url)
	time.Sleep(time.Second)
	if e != nil {
		r.Body.Close()
		panic("WTF")
	}
}
