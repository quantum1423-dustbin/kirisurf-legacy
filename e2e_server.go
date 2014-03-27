// e2e_server.go
package main

import (
	//"io"
	"net"
	"time"

	"github.com/coreos/go-log/log"
)

// e2e server handler. Subcircuit calls this.
func e2e_server_handler(wire *gobwire) {
	chantable := make(map[int]chan e2e_segment)
	die_pl0x := false
	defer wire.destroy()
	defer log.Debug("Exiting...")
	defer func() { die_pl0x = true }()
	defer func() {
		for _, ch := range chantable {
			if ch != nil {
				ch <- e2e_segment{E2E_CLOSE, 0, []byte("")}
			}
		}
	}()
	for {
		if die_pl0x {
			log.Debug("die_pl0x received, die!")
			return
		}
		log.Debug("waiting for new seg...")
		newseg, err := wire.Receive()
		if err != nil {
			log.Debug("e2e server handler encountered error: ", err.Error())
			return
		}
		if newseg.Flag == E2E_OPEN {
			log.Debug("e2e server received open connection request, Connid ", newseg.Connid)
			commchan := make(chan e2e_segment)
			chantable[newseg.Connid] = commchan
			go func() {
				commchan := commchan
				Connid := newseg.Connid
				// prepare in case of emergencyings
				close_pkt := e2e_segment{E2E_CLOSE, newseg.Connid, []byte("byeings")}
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
					defer log.Debug("remote is of closed")
					defer rmt.Close()
					for {
						if die_pl0x {
							log.Debug("I guess we should die now...")
							return
						}
						pkt := <-commchan
						if pkt.Flag == E2E_CLOSE {
							log.Debug("E2E_CLOSE received")
							return
						}
						_, err := rmt.Write(pkt.Body)
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
					data_pkt := e2e_segment{E2E_DATA, Connid, thing}
					err = wire.Send(data_pkt)
					if err != nil {
						log.Debug("Write error to client: ", err.Error())
						die_pl0x = true
						return
					}
				}
			}()
		} else if newseg.Flag == E2E_DATA || newseg.Flag == E2E_CLOSE {
			Connid := newseg.Connid
			if chantable[Connid] == nil {
				log.Debug("Connid of nil? Pl0x die!")
				die_pl0x = true
				continue
			}
			chantable[Connid] <- newseg
		} else {
			log.Error("e2e server received a packet with a weird Flag")
			die_pl0x = true
			return
		}
	}
}
