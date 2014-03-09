// kirisurf project main.go
package main

import (
<<<<<<< HEAD
	"encoding/base32"
	"kirisurf/subcircuit"
	"libkiricrypt"
	"libkiridir"
	"libkiss"
	"runtime"
	"strconv"
	"strings"
=======
	"fmt"
	"libkirill"
	"runtime"
>>>>>>> 6a6961b7121a0a72f8c9e682664fcf7fa23ed4cc

	"code.google.com/p/log4go"
)

<<<<<<< HEAD
var MasterKey = libkiricrypt.SecureDH_genpair()
var MasterKeyHash = strings.ToLower(base32.StdEncoding.EncodeToString(
	libkiricrypt.InvariantHash(MasterKey.Public.Bytes())[:20]))

func main() {
	log4go.Info("Kirisurf starting; MasterKeyHash=%s", MasterKeyHash)
	libkiss.SetCipher(libkiricrypt.AS_blowfish128_ofb)
	runtime.GOMAXPROCS(runtime.NumCPU())
	ourport, _ := (strconv.Atoi(strings.Split(MasterConfig.General.ORAddr, ":")[1]))
	go libkiridir.RunRelay(ourport, MasterKeyHash, MasterConfig.General.IsExit)
	subcircuit.TestServer()
	log4go.Info("Kirisurf exited.")
=======
func main() {
	libkirill.SetCipher(libkirill.AS_blowfish128_ofb)
	//libkirill.KIRILL_DEBUG = false
	log4go.Debug("Hello world")
	//libkiridir.RunRelay(20123, "sjdfklsdjf", false)
	libkirill.LOG(libkirill.LOG_DEBUG, "FJKLJASD")
	//libkiridir.RefreshDirectory()
	runtime.GOMAXPROCS(runtime.NumCPU())
	libkirill.KiSS_test()
	fmt.Println("Kirisurf exited.")
>>>>>>> 6a6961b7121a0a72f8c9e682664fcf7fa23ed4cc
}
