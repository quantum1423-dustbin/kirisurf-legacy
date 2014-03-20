package main

import (
	"code.google.com/p/gcfg"
	"github.com/coreos/go-log/log"
)

type Config struct {
	General struct {
		Role         string
		SocksAddr    string
		ORAddr       string
		IsExit       bool
		DirectoryURL string
	}
	Network struct {
		MinCircuitLen int
	}
}

var MasterConfig Config

func init() {
	err := gcfg.ReadFileInto(&MasterConfig, "kirisurf.conf")
	log.Debug("Read configuration successfully")
	if err != nil {
		panic(err.Error())
	}
}
