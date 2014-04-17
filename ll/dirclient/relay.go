// relay.go
package dirclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	PROTVER = 200
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
	for {
		time.Sleep(time.Second)
		r, e := http.Get(fmt.Sprintf("%s/longpoll", DIRADDR))
		if e != nil {
		retry:
			url := fmt.Sprintf("%s/upload?port=%d&protocol=%d&keyhash=%s&exit=%d",
				DIRADDR,
				port, PROTVER, keyhash, ieflag)
			_, e := http.Get(url)
			if e != nil {
				goto retry
			}
			continue
		}
		buff := new(bytes.Buffer)
		io.Copy(buff, r.Body)
		protector.Lock()
		err := json.Unmarshal(buff.Bytes(), &KDirectory)
		protector.Unlock()
		if err != nil {
			r.Body.Close()
		retryy:
			url := fmt.Sprintf("%s/upload?port=%d&protocol=%d&keyhash=%s&exit=%d",
				DIRADDR,
				port, PROTVER, keyhash, ieflag)
			_, e := http.Get(url)
			if e != nil {
				goto retryy
			}
			continue
		}
	}
}
