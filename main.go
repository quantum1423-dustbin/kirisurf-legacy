// kirisurf project main.go
package main

import (
	"flag"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/kiss"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var MasterKey = kiss.GenerateDHKeys()
var MasterKeyHash = hash_base32(MasterKey.Public)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

var version = "NOT_A_RELEASE_VERSION"

func main() {

	INFO("Kirisurf %s started! CPU count: %d, mkh=%s", version, runtime.NumCPU(), MasterKeyHash)
	runtime.GOMAXPROCS(runtime.NumCPU() * 3)
	set_gui_progress(0.1)
	INFO("Bootstrapping 10%%: finding directory address...")
	dirclient.DIRADDR, _ = dirclient.FindDirectoryURL()
	set_gui_progress(0.2)
	INFO("Bootstrapping 20%%: found directory address, refreshing directory...")
	err := dirclient.RefreshDirectory()
	if err != nil {
		CRITICAL("Stuck at 20%%: directory connection error %s", err.Error())
		for {
			time.Sleep(time.Second)
		}
	}
	set_gui_progress(0.3)
	INFO("Bootstrapping 30%%: directory refreshed, beginning to build circuits...")
	INFO(MasterConfig.General.Role)
	go run_diagnostic_loop()
	dirclient.RefreshDirectory()
	if MasterConfig.General.Role == "server" {
		NewSCServer(MasterConfig.General.ORAddr)
		prt, _ := strconv.Atoi(
			strings.Split(MasterConfig.General.ORAddr, ":")[1])
		go dirclient.RunRelay(prt, MasterKeyHash,
			MasterConfig.General.IsExit)
		INFO("Bootstrapping 100%%: server started!")
		for {
			time.Sleep(time.Second * 10)
		}
	}
	run_client_loop()
	INFO("Kirisurf exited")
}
