package main

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
	MasterConfig.Network.MinCircuitLen = 4
	MasterConfig.General.SocksAddr = "127.0.0.1:9090"
	MasterConfig.General.IsExit = false
	MasterConfig.General.ORAddr = "OMGOMGOMDONOTUSEATALL"
}
