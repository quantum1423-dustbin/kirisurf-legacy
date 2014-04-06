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
	go run_diagnostic_loop()
	kiss.SetCipher(kicrypt.AS_blowfish448_ofb)
	//kiss.KiSS_test()
	dirclient.DIRADDR = MasterConfig.General.DirectoryURL
	log.Info("Kirisurf started")
	dirclient.RefreshDirectory()
	runtime.GOMAXPROCS(runtime.NumCPU())
	if MasterConfig.General.Role == "server" {
		go socks5proxy()
		bigserve := NewSCServer(MasterConfig.General.ORAddr)
		prt, _ := strconv.Atoi(
			strings.Split(MasterConfig.General.ORAddr, ":")[1])
		dirclient.RunRelay(prt, MasterKeyHash,
			MasterConfig.General.IsExit)
		for {
			time.Sleep(time.Second)
		}
		bigserve.Kill()
	} else if MasterConfig.General.Role == "client" {
		run_client_loop()
	}
	log.Info("Kirisurf exited")
}
