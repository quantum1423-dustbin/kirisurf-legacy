// e2e_server.go
package main

import (
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/coreos/go-log/log"
)

// e2e server handler. Subcircuit calls this.
func e2e_server_handler(wire *gobwire) {
	// KILLSWITCH. Use in select, when this thing closes everything is going to die
	KILLSWITCH := make(chan bool)

	chantable := make(map[int]chan e2e_segment)
	var tablock sync.RWMutex
	conntable := make(map[int]net.Conn)
	// global upstream and downstream
	gupstream := make(chan e2e_segment, 16)
	gdownstream := make(chan e2e_segment, 16)
	IAMDEAD := false
	var once sync.Once
	global_die := func() {
		once.Do(func() {
			IAMDEAD = true
			log.Debug("global_die() called!")
			log.Debug("signalling KILLSWITCH")
			close(KILLSWITCH)
			log.Debug("sleeping 1 second to prevent race...")
			time.Sleep(time.Second)
			log.Debug("destroying global objects...")
			close(gupstream)
			close(gdownstream)
			wire.destroy()
			for _, ch := range chantable {
				if ch != nil {
					close(ch)
				}
			}
			log.Debug("chantable closed")
			chantable = nil
			log.Debug("chantable nilled")
			for _, cn := range conntable {
				if cn != nil {
					cn.Close()
				}
			}
			log.Debug("conntable closed")
			conntable = nil
			log.Debug("conntable nilled")
			log.Debug("collecting garbage and exiting...")
			debug.FreeOSMemory()
		})
	}

	// goroutines for makings of stronk up and down
	go func() {
		defer wire.destroy()
		for {
			newpkt, ok := <-gdownstream
			if !ok {
				log.Debug("gdownstream got not ok, dying")
				global_die()
				return
			}
			err := wire.Send(newpkt)
			if err != nil {
				log.Debug("gdownstream send error: ", err.Error())
				log.Debug("gdownstream dying")
				global_die()
				return
			}
		}
	}()
	go func() {
		defer wire.destroy()
		for {
			newpkt, err := wire.Receive()
			if err != nil {
				log.Debug("gupstream receive error, closing")
				global_die()
				return
			}
			gupstream <- newpkt
		}
	}()
	defer wire.conn.Close()
	for {
		select {
		case <-KILLSWITCH:
			log.Debug("KILLSWITCH received in main loop, returning")
			return
		case thing, ok := <-gupstream:
			if !ok {
				log.Debug("gupstream not of okays, dying")
				global_die()
				return
			}
			if thing.Flag == E2E_OPEN {
				connid := thing.Connid
				log.Debugf("E2E_OPEN(%d)", connid)
				chantable[connid] = make(chan e2e_segment, 16)
				go func() {
					conn, err := net.DialTimeout("tcp", SOCKSADDR, time.Second*20)
					closepak := e2e_segment{E2E_CLOSE, connid, []byte("")}
					defer func() {
						if !IAMDEAD {
							gdownstream <- closepak
						}
					}()
					tablock.Lock()
					conntable[connid] = conn
					ch := chantable[connid]
					tablock.Unlock()
					if err != nil {
						log.Debug("Error encountered in remote: ", err.Error())
						return
					}
					defer conn.Close()

					// Token bucket
					tokenbucket := make(chan bool, 1024) // 4mb grace period
					for i := 0; i < 1000; i++ {
						tokenbucket <- true
					}
					hardtb := make(chan bool, 256) // Hard tb for sendme coordination
					go func() {
						for {
							select {
							case tokenbucket <- true:
								time.Sleep(time.Second / 10)
							case <-KILLSWITCH:
								return
							}
						}
					}()

					// Downstream
					go func() {
						defer conn.Close()
						for {
							buf := make([]byte, 4096)
							select {
							case <-KILLSWITCH:
								log.Debug("KILLSWITCH signalled on remote downstr!")
								return
							default:
								// Obtain token
								<-hardtb
								<-tokenbucket
								n, err := conn.Read(buf)
								if err != nil {
									log.Debug("Received error from remote")
									return
								}
								ah := make([]byte, n)
								copy(ah, buf[:n])
								tosend := e2e_segment{E2E_DATA, connid, ah}
								gdownstream <- tosend
							}
						}
					}()
					// Upstream
					for {
						select {
						case <-KILLSWITCH:
							log.Debug("KILLSWITCH signalled on remote conn!")
							return
						case newthing, ok := <-ch:
							if !ok {
								log.Debug("connection chan closed!")
								return
							}
							if newthing.Flag == E2E_CLOSE {
								log.Debug("E2E_CLOSE received")
								return
							}
							if newthing.Flag == E2E_SENDMORE {
								log.Debug("Send more...")
								for i := 0; i < 256; i++ {
									select {
									case hardtb <- true:
									default:
									}
								}
							}
							_, err := conn.Write(newthing.Body)
							if err != nil {
								log.Debug("Error while writing to remote: ", err.Error())
								return
							}
						}
					}
				}()
			} else if thing.Flag == E2E_DATA || thing.Flag == E2E_CLOSE {
				tablock.RLock()
				ch := chantable[thing.Connid]
				tablock.RUnlock()
				ch <- thing
			} else {
				log.Debug("Weird, weird, weird segment received!")
				global_die()
				return
			}
		}
	}
}
