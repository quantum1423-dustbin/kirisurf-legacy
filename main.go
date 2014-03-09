// kirisurf project main.go
package main

import (
	"encoding/base32"
	"kirisurf/subcircuit"
	"libkiricrypt"
	"libkiridir"
	"libkiss"
	"runtime"
	"strconv"
	"strings"

	"code.google.com/p/log4go"
)

var MasterKey = libkiricrypt.SecureDH_genpair()
var MasterKeyHash = strings.ToLower(base32.StdEncoding.EncodeToString(
	libkiricrypt.InvariantHash(MasterKey.Public.Bytes())[:20]))

func main() {
	log4go.Info("Kirisurf starting; MasterKeyHash=%s", MasterKeyHash)
	libkiss.SetCipher(libkiricrypt.AS_blowfish128_ofb)
	runtime.GOMAXPROCS(runtime.NumCPU())
	ourport, _ := (strconv.Atoi(strings.Split(MasterConfig.General.ORAddr, ":")[1]))
	go libkiridir.RunRelay(ourport, MasterKeyHash, MasterConfig.General.IsExit)
	subcircuit.TestServer()
	log4go.Info("Kirisurf exited.")
}
