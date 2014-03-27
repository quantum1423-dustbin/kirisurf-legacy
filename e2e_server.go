// e2e_server.go
package main

import (
	"io"
	"net"
	"time"

	"github.com/coreos/go-log/log"
)

// e2e server handler. Subcircuit calls this.
func e2e_server_handler(client io.ReadWriteCloser) {
	wire := newGobWire(client)
	chantable := make(map[int]chan e2e_segment)
	die_pl0x := false
	defer wire.destroy()
	for {
		if die_pl0x {
			log.Debug("die_pl0x received, die!")
			return
		}
		newseg, err := wire.Receive()
		if err != nil {
			log.Debug("e2e server handler encountered error: ", err.Error())
			return
		}
		if newseg.flag == E2E_OPEN {
			log.Debug("e2e server received open connection request, connid ", newseg.connid)
			commchan := make(chan e2e_segment)
			chantable[newseg.connid] = commchan
			go func() {
				commchan := commchan
				connid := newseg.connid
				// prepare in case of emergencyings
				close_pkt := e2e_segment{E2E_CLOSE, newseg.connid, []byte("byeings")}
				defer wire.Send(close_pkt)
				// open das connektion
				rmt, err := net.DialTimeout("tcp", SOCKSADDR, time.Second*20)
				if err != nil {
					log.Debug("e2e server encountered forwarding error: ", err.Error())
					return
				}
				defer rmt.Close()
				// fork out the upstream handler
				go func() {
					defer rmt.Close()
					for {
						if die_pl0x {
							log.Debug("I guess we should die now...")
							return
						}
						pkt := <-commchan
						if pkt.flag == E2E_CLOSE {
							log.Debug("E2E_CLOSE received")
							return
						}
						_, err := rmt.Write(pkt.body)
						if err != nil {
							log.Debug("Write error to remote: ", err.Error())
							return
						}
					}
				}()
				// downstream
				buf := make([]byte, 16384)
				for {
					if die_pl0x {
						log.Debug("I guess we should die now...")
						return
					}
					n, err := rmt.Read(buf)
					if err != nil {
						log.Debug("Read error from remote: ", err.Error())
						return
					}
					thing := make([]byte, n)
					copy(thing, buf[:n])
					data_pkt := e2e_segment{E2E_DATA, connid, thing}
					err = wire.Send(data_pkt)
					if err != nil {
						log.Debug("Write error to client: ", err.Error())
						die_pl0x = true
						return
					}
				}
			}()
		} else if newseg.flag == E2E_DATA || newseg.flag == E2E_CLOSE {
			connid := newseg.connid
			if chantable[connid] == nil {
				die_pl0x = true
				continue
			}
			chantable[connid] <- newseg
		} else {
			log.Error("e2e server received a packet with a weird flag")
			die_pl0x = true
			return
		}
	}
}
