// kirisurf project main.go
package main

import (
	"encoding/base32"
	"libkiricrypt"
	"libkiridir"
	"libkiss"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-log/log"
)

var MasterKey = libkiricrypt.SecureDH_genpair()
var MasterKeyHash = strings.ToLower(base32.StdEncoding.EncodeToString(
	libkiricrypt.InvariantHash(MasterKey.Public.Bytes())[:20]))

func main() {
	log.Info("Kirisurf started")
	libkiridir.RefreshDirectory()
	libkiss.SetCipher(libkiricrypt.AS_blowfish128_ofb)
	runtime.GOMAXPROCS(runtime.NumCPU())
	if MasterConfig.General.Role == "server" {
		bigserve := NewSCServer(MasterConfig.General.ORAddr)
		prt, _ := strconv.Atoi(
			strings.Split(MasterConfig.General.ORAddr, ":")[1])
		libkiridir.RunRelay(prt, MasterKeyHash,
			MasterConfig.General.IsExit)
		for {
			time.Sleep(time.Second)
		}
		bigserve.Kill()
	}
	log.Info("Kirisurf exited")
}
