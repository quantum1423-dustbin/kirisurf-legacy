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
	for {
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
		if e != nil {
			time.Sleep(time.Second * 10)
			continue
		}
		r.Body.Close()
		time.Sleep(time.Second * 500)
	}
}
