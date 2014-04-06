// libkiridir project libkiridir.go
package dirclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

var DIRADDR = "https://directory.kirisurf.org/"

// Maintains the big database
type KNode struct {
	PublicKey       string
	Address         string
	ProtocolVersion int
	Adjacents       []int
	ExitNode        bool
}

// The database itself
var KDirectory = make([]KNode, 0)

// lock
var protector sync.RWMutex

// Refresh the directory
func RefreshDirectory() error {
	protector.Lock()
	defer protector.Unlock()
	resp, err := http.Get(strings.Join([]string{DIRADDR, "read"}, ""))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buff := new(bytes.Buffer)
	io.Copy(buff, resp.Body)
	err = json.Unmarshal(buff.Bytes(), &KDirectory)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(KDirectory)
	return nil
}

// Lookup based on a public key
func PKeyLookup(pkey string) *KNode {
	protector.RLock()
	defer protector.RUnlock()
	toret := (*KNode)(nil)
	for _, val := range KDirectory {
		if val.PublicKey == pkey {
			toret = &val
			break
		}
	}
	return toret
}
