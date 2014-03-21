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
	libkiss.SetCipher(libkiricrypt.AS_blowfish128_ofb)
	//libkiss.KiSS_test()
	log.Info("Kirisurf started")
	libkiridir.RefreshDirectory()
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
	sc, err := build_subcircuit()
	if err != nil {
		panic(err.Error())
	}
	sc.wire.Close()
	log.Info("Kirisurf exited")
}
