// relay.go
package dirclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
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
	log.Debug(url)
	time.Sleep(time.Second)
	if e != nil {
		log.Errorf("Error encountered in info upload: %s", e.Error())
		r.Body.Close()
		panic("WTF")
	}
	for {
		r, e := http.Get("https://directory.kirisurf.org/longpoll")
		time.Sleep(time.Second)
		if e != nil {
			log.Errorf("Error encountered in long poll: %s", e.Error())
			//r.Body.Close()
			continue
		}
		buff := new(bytes.Buffer)
		io.Copy(buff, r.Body)
		protector.Lock()
		err := json.Unmarshal(buff.Bytes(), &KDirectory)
		protector.Unlock()
		if err != nil {
			log.Errorf("Error encountered when decoding long poll: %s / %s",
				err.Error(), string(buff.Bytes()))
			r.Body.Close()
			continue
		}
	}
}
