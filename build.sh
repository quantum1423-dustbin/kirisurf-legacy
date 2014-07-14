#!/bin/bash
go build -ldflags "-X main.version $(git describe --tags --long)" -v
