// kirisurf project main.go
package main

import (
	"fmt"
	"libkiridir"
	"libkirill"
	"runtime"

	"code.google.com/p/log4go"
)

func main() {
	libkirill.KIRILL_DEBUG = false
	log4go.Debug("Hello world")
	libkiridir.RunRelay(20123, "sjdfklsdjf", false)
	libkirill.LOG(libkirill.LOG_DEBUG, "FJKLJASD")
	libkiridir.RefreshDirectory()
	runtime.GOMAXPROCS(runtime.NumCPU())
	libkirill.KiSS_test()
	fmt.Println("Kirisurf exited.")
}
