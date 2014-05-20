package main

import (
	"encoding/json"
	"kirisurf/ll/dirclient"
	"net/http"
)

type TrafficStats struct {
	DownloadedBytes int
	UploadedBytes   int
}

// Shortpoll for traffic
func webface_traffic_json(w http.ResponseWriter, req *http.Request) {
	// Traffic indicator
	global_locker.Lock()
	downbts := global_down_bytes
	upbts := global_up_bytes
	global_locker.Unlock()

	var towrite TrafficStats
	towrite.DownloadedBytes = downbts
	towrite.UploadedBytes = upbts

	res, err := json.MarshalIndent(&towrite, "", "\t")
	if err != nil {
		return
	}
	w.Write(res)
}

type DirectoryInfo struct {
	Directory        []dirclient.KNode
	PossibleCircuits [][]dirclient.KNode
}

// Shortpoll for directory info
func webface_directoryinfo_json(w http.ResponseWriter, req *http.Request) {
	var xaxa DirectoryInfo
	xaxa.Directory = dirclient.KDirectory
	xaxa.PossibleCircuits = viableNodes

	res, err := json.MarshalIndent(&xaxa, "", "\t")
	if err != nil {
		return
	}
	w.Write(res)
}
