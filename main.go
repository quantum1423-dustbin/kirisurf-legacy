// kirisurf project main.go
package main

import (
	"flag"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/kiss"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"code.google.com/p/gcfg"

	"github.com/KirisurfProject/kilog"
)

var MasterKey = kiss.GenerateDHKeys()
var MasterKeyHash = hash_base32(MasterKey.Public)

var confloc = flag.String("c", "", "config location")
var singhop = flag.Bool("singhop", false, "single hop or not")

var version = "NOT_A_RELEASE_VERSION"

func main() {
	rand.Seed(time.Now().UnixNano())
	go run_monitor_loop()
	flag.Parse()
	if *confloc == "" {
		kilog.Warning("No configuration file given, using defaults")
	} else {
		err := gcfg.ReadFileInto(&MasterConfig, *confloc)
		if err != nil {
			kilog.Warning("Configuration file broken, using defaults")
		}
	}
	if *singhop {
		MasterConfig.Network.MinCircuitLen = 1
	}
	INFO("Kirisurf %s started! mkh=%s", version, MasterKeyHash)
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
		RegisterNGSCServer(MasterConfig.General.ORAddr)
		prt, _ := strconv.Atoi(
			strings.Split(MasterConfig.General.ORAddr, ":")[1])
		go dirclient.RunRelay(prt, MasterKeyHash,
			MasterConfig.General.IsExit)
		set_gui_progress(1.0)
		INFO("Bootstrapping 100%%: server started!")
		for {
			time.Sleep(time.Second * 10)
		}
	}
	run_client_loop()
	INFO("Kirisurf exited")
}
