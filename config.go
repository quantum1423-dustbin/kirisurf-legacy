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
	err := gcfg.ReadFileInto(&MasterConfig, "kirisurf.conf")
	if err != nil {
		panic(err.Error())
	}
}
