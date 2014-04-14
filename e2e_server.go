// e2e_server.go
package main

import (
	"net"
	"sync"
	"time"
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
			DEBUG("global_die() called on %p!", wire)
			DEBUG("signalling KILLSWITCH on %p!", wire)
			close(KILLSWITCH)
			DEBUG("sleeping 1 second on %p to prevent race...", wire)
			time.Sleep(time.Second)
			DEBUG("destroying global objects for %p...", wire)
			close(gupstream)
			close(gdownstream)
			wire.destroy()
			for _, ch := range chantable {
				if ch != nil {
					close(ch)
				}
			}
			for _, cn := range conntable {
				if cn != nil {
					cn.Close()
				}
			}
			DEBUG("exiting from %p...", wire)
		})
	}

	// goroutines for makings of stronk up and down
	go func() {
		defer wire.destroy()
		for {
			newpkt, ok := <-gdownstream
			if !ok {
				global_die()
				return
			}
			err := wire.Send(newpkt)
			if err != nil {
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
			DEBUG("KILLSWITCH received in main loop for %p, returning", wire)
			return
		case thing, ok := <-gupstream:
			if !ok {
				DEBUG("gupstream not of okays, dying (%p)", wire)
				global_die()
				return
			}
			if thing.Flag == E2E_OPEN {
				connid := thing.Connid
				chantable[connid] = make(chan e2e_segment, 16)
				go func() {
					DEBUG("Connection request to %s (%p)", string(thing.Body), wire)
					conn, err := net.DialTimeout("tcp", string(thing.Body), time.Second*20)
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
						DEBUG("Error encountered in remote (%p): %s", wire, err.Error())
						return
					}
					defer conn.Close()

					// Token bucket
					tokenbucket := make(chan bool, 1024) // 4mb grace period
					for i := 0; i < 1000; i++ {
						tokenbucket <- true
					}
					hardtb := make(chan bool, 256) // Hard tb for sendme coordination
					for i := 0; i < 256; i++ {
						select {
						case hardtb <- true:
						default:
						}
					}
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
								DEBUG("KILLSWITCH signalled on remote downstr! (%p)", wire)
								return
							default:
								// Obtain token
								<-hardtb
								<-tokenbucket
								n, err := conn.Read(buf)
								if err != nil {
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
							DEBUG("KILLSWITCH signalled on remote conn! (%p)", wire)
							return
						case newthing, ok := <-ch:
							if !ok {
								return
							}
							if newthing.Flag == E2E_CLOSE {
								return
							}
							if newthing.Flag == E2E_SENDMORE {
								DEBUG("Sending more for [%p]:%d", wire, connid)
								for i := 0; i < 256; i++ {
									select {
									case hardtb <- true:
									default:
									}
								}
							}
							_, err := conn.Write(newthing.Body)
							if err != nil {
								DEBUG("Error while writing to remote (%p): %s", wire, err.Error())
								return
							}
						}
					}
				}()
			} else if thing.Flag == E2E_DATA || thing.Flag == E2E_CLOSE || thing.Flag == E2E_SENDMORE {
				tablock.RLock()
				ch := chantable[thing.Connid]
				tablock.RUnlock()
				ch <- thing
			} else {
				global_die()
				return
			}
		}
	}
}
