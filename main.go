// kirisurf project main.go
package main

import (
	"encoding/base32"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/kicrypt"
	"kirisurf/ll/kiss"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-log/log"
)

var MasterKey = kicrypt.SecureDH_genpair()
var MasterKeyHash = strings.ToLower(base32.StdEncoding.EncodeToString(
	kicrypt.InvariantHash(MasterKey.Public.Bytes())[:20]))

func main() {
	kiss.SetCipher(kicrypt.AS_blowfish128_ofb)
	//kiss.KiSS_test()
	log.Info("Kirisurf started")
	dirclient.RefreshDirectory()
	runtime.GOMAXPROCS(runtime.NumCPU())
	if MasterConfig.General.Role == "server" {
		bigserve := NewSCServer(MasterConfig.General.ORAddr)
		prt, _ := strconv.Atoi(
			strings.Split(MasterConfig.General.ORAddr, ":")[1])
		dirclient.RunRelay(prt, MasterKeyHash,
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
