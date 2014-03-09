package main

import "code.google.com/p/gcfg"

type Config struct {
	General struct {
		SocksAddr    string
		ORAddr       string
		IsExit       bool
		DirectoryURL string
	}
}

var MasterConfig Config

func init() {
	err := gcfg.ReadFileInto(&MasterConfig, "kirisurf.conf")
	if err != nil {
		panic(err.Error())
	}
}
