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
var noclient = flag.Bool("noclient", false, "server only")

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
	kilog.Info("Kirisurf %s started! mkh=%s", version, MasterKeyHash)
	set_gui_progress(0.1)
	kilog.Info("Bootstrapping 10%%: finding directory address...")
	dirclient.DIRADDR, _ = dirclient.FindDirectoryURL()
	set_gui_progress(0.2)
	kilog.Info("Bootstrapping 20%%: found directory address, refreshing directory...")
	err := dirclient.RefreshDirectory()
	if err != nil {
		kilog.Critical("Stuck at 20%%: directory connection error %s", err.Error())
		for {
			time.Sleep(time.Second)
		}
	}
	set_gui_progress(0.3)
	kilog.Info("Bootstrapping 30%%: directory refreshed, beginning to build circuits...")
	kilog.Info(MasterConfig.General.Role)
	go run_diagnostic_loop()
	dirclient.RefreshDirectory()
	if MasterConfig.General.Role == "server" {
		NewSCServer(MasterConfig.General.ORAddr)
		addr := RegisterNGSCServer(MasterConfig.General.ORAddr)
		prt, _ := strconv.Atoi(
			strings.Split(MasterConfig.General.ORAddr, ":")[1])
		go func() {
			err := UPnPForwardAddr(MasterConfig.General.ORAddr)
			if err != nil {
				kilog.Warning("UPnP failed: %s", err)
				if MasterConfig.Network.OverrideUPnP {
					go dirclient.RunRelay(prt, MasterKeyHash,
						MasterConfig.General.IsExit)
				}
				return
			}
			err = UPnPForwardAddr(addr)
			if err != nil {
				kilog.Warning("UPnP failed: %s", err)
				if MasterConfig.Network.OverrideUPnP {
					go dirclient.RunRelay(prt, MasterKeyHash,
						MasterConfig.General.IsExit)
				}
				return
			}
			kilog.Info("UPnP successfully forwarded port")
			go dirclient.RunRelay(prt, MasterKeyHash,
				MasterConfig.General.IsExit)
		}()
		kilog.Info("Started server!")
		if *noclient {
			for {
				time.Sleep(time.Second)
			}
		}
	}
	run_client_loop()
	kilog.Info("Kirisurf exited")
}
