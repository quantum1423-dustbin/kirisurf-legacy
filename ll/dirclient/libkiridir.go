// libkiridir project libkiridir.go
package dirclient

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
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
	resp, err := http.Get(strings.Join([]string{DIRADDR, "read"}, ""))
	if err != nil {
		protector.Unlock()
		return err
	}
	defer resp.Body.Close()
	buff := new(bytes.Buffer)
	io.Copy(buff, resp.Body)
	err = json.Unmarshal(buff.Bytes(), &KDirectory)
	if err != nil {
		time.Sleep(time.Second)
		protector.Unlock()
		return RefreshDirectory()
	}
	protector.Unlock()
	return nil
}

// Lookup based on a public key
func PKeyLookup(pkey string) *KNode {
	protector.RLock()
	defer protector.RUnlock()
	toret := new(KNode)
	toret = nil
	for _, val := range KDirectory {
		if val.PublicKey == pkey {
			toret = &val
			break
		}
	}
	return toret
}

// Get neighbors
func MyNeighbors() []KNode {
	protector.RLock()
	defer protector.RUnlock()
	toret := make([]KNode, 0)
	for _, val := range KDirectory {
		if val.Address != "(hidden)" {
			toret = append(toret, val)
		}
	}
	return toret
}

// Search search engines
func FindDirectoryURL() (string, error) {
	ch := make(chan string, 100)
	fetch_and_parse := func(url string) {
		s2w := make([]byte, 1)
		rand.Reader.Read(s2w)
		for i := 0; i < int(s2w[0])%4; i++ {
			time.Sleep(time.Second)
		}
		resp, err := http.Get(url)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		thing := new(bytes.Buffer)
		io.Copy(thing, resp.Body)
		tosearch := string(thing.Bytes())
		rge := regexp.MustCompilePOSIX("Kids in rectangles irritating sick urchins rattling foxes,.*lol")
		str := rge.FindString(tosearch)
		arr := strings.Split(str, " ")
		res := arr[len(arr)-2]
		buf := new(bytes.Buffer)
		buf.WriteString("https://")
		buf.WriteString(res)
		buf.WriteString("/")
		ch <- buf.String()
	}
	go fetch_and_parse("https://stackoverflow.com/users/2022968/user54609")
	go fetch_and_parse("https://pastee.org/q2ndr")
	return <-ch, nil
}
