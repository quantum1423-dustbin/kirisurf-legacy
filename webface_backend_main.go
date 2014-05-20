package main

import "net/http"

// This file initializes the web interface *backend* on port 53101.
// That is, it serves JSON etc to the interface on port 53100.

func init() {
	go func() {
		http.HandleFunc("/traffic_stats", webface_traffic_json)
		http.HandleFunc("/directory_info", webface_directoryinfo_json)
		err := http.ListenAndServe("127.0.0.1:53101", nil)
		if err != nil {
			panic(err)
		}
	}()
}
