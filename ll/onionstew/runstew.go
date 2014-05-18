package onionstew

import (
	"fmt"

	"github.com/KirisurfProject/kilog"
)

func (ctx *stew_ctx) run_stew(is_server bool) {
	defer func() {
		if x := recover(); x != nil {
			kilog.Debug("%v", x)
		}
	}()
	for {
		select {
		case <-ctx.killswitch:
			kilog.Debug("KILLSWITCH signalled for stew layer!")
			return
		case thing := <-ctx.llctx.ordered_ch:
			//kilog.Debug("Received thing with ctr=%d (%v)", thing.seqnum, is_server)
			pkt := bytes_to_stew_message(thing.payload)
			if pkt.category == m_open && is_server {
				remote_addr := string(pkt.payload)
				desired_connid := pkt.connid
				ctx.lock.Lock()
				ctx.conntable[desired_connid] = make(chan stew_message, 256)
				ctx.lock.Unlock()
				go ctx.attacht_remote(remote_addr, desired_connid)
			} else if pkt.category == m_close || pkt.category == m_data || pkt.category == m_more {
				ctx.lock.RLock()
				ch := ctx.conntable[pkt.connid]
				ctx.lock.RUnlock()
				if ch == nil {
					kilog.Debug("stew_message with illegal connid received, ignoring")
					continue
				}
				ch <- pkt
			} else if pkt.category == m_dns {
				// not implemented
			} else {
				// illegal packet! Log & silent-drop as a precaution against attakk
				kilog.Warning("Illegal stew_message received: %s, %x (is_server=%b)",
					fmt.Sprint(pkt), pkt.bytes(), is_server)
			}
		case client := <-ctx.client_ch:
			if is_server {
				panic("Why the hell did you write to client_ch as a SERVER?!?!?")
			}
			go ctx.attacht_client(client)
		}
	}
}
