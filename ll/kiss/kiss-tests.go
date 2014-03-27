package kiss

import (
	//"fmt"
	"kirisurf/ll/kicrypt"
	"net"
	//"os"
	"time"
)

func KiSS_test() {
	server_key := kicrypt.SecureDH_genpair()
	server_with_dispatch("localhost:5555",
		func(owire net.Conn) {
			//owire, err := Kiriobfs_handshake_server(wire)
			//check_serious(err)
			wrapped, err := KiSS_handshake_server(owire, server_key)
			wrapped, err = KiSS_handshake_server(wrapped, server_key)
			wrapped, err = KiSS_handshake_server(wrapped, server_key)
			check_serious(err)
			LOG(LOG_DEBUG, "Accepted...")
			copy_conns(wrapped, wrapped)
		})

	server_with_dispatch("0.0.0.0:6666",
		func(wire net.Conn) {
			remwire, err := net.Dial("tcp", "localhost:5555")
			check_serious(err)
			//remwire, err := Kiriobfs_handshake_client(gremwire)
			//check_serious(err)
			remreal, err1 := KiSS_handshake_client(remwire, dumb_Verifier)
			remreal, err1 = KiSS_handshake_client(remreal, dumb_Verifier)
			remreal, err1 = KiSS_handshake_client(remreal, dumb_Verifier)
			check_serious(err1)
			LOG(LOG_DEBUG, "Connected...")
			go copy_conns(wire, remreal)
			copy_conns(remreal, wire)
		})

	time.Sleep(10000 * time.Second)
}
