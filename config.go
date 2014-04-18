package main

import "code.google.com/p/gcfg"

type Config struct {
	General struct {
		Role      string
		SocksAddr string
		ORAddr    string
		IsExit    bool
	}
	Network struct {
		MinCircuitLen int
	}
}

var MasterConfig Config

func init() {
	//defaults
	MasterConfig.General.Role = "client"
	MasterConfig.Network.MinCircuitLen = 5
	MasterConfig.General.SocksAddr = "127.0.0.1:9090"
	MasterConfig.General.IsExit = false
	MasterConfig.General.ORAddr = "OMGOMGOMDONOTUSEATALL"
	err := gcfg.ReadFileInto(&MasterConfig, "kirisurf.conf")
	if err != nil {
		WARNING("*** Config file broken (%s), using defaults ***", err.Error())
	}
}
