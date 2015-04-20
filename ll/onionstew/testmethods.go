package onionstew

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/KirisurfProject/kilog"
)

func test_socks_client() {
	return /*
		generator := func() io.ReadWriteCloser {
			toret, err := net.Dial("tcp", "oozora.servers.kirisurf-legacy.org:5555")
			if err != nil {
				panic(err.Error())
			}
			return toret
		}
		haha, err := MakeManagedClient(generator)
		if err != nil {
			panic(err.Error())
		}
		listener, err := net.Listen("tcp", "127.0.0.1:6666")
		if err != nil {
			panic(err.Error())
		}
		for {
			thing, err := listener.Accept()
			if err != nil {
				continue
			}
			haha.AddClient(thing)
		}*/
}

func RunManagedStewServer() string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err.Error())
	}

	// Keep trakk of shteewz
	type stid struct {
		top uint64
		bot uint64
	}
	stew_table := make(map[stid]*stew_ctx)
	var stew_table_lk sync.Mutex
	kilog.Debug("before gc goroutine")
	// Garbage collect stew table every 30 seconds
	go func() {
		for {
			time.Sleep(time.Second * 30)
			stew_table_lk.Lock()
			for stid, ctx := range stew_table {
				ctx.llctx.lock.Lock()
				cnt := ctx.llctx.refcount
				ctx.llctx.lock.Unlock()
				if cnt == 0 {
					ctx.destroy()
					stid := stid
					go func() {
						stew_table_lk.Lock()
						kilog.Debug("Collecting %d:%d...", stid.top, stid.bot)
						delete(stew_table, stid)
						stew_table_lk.Unlock()
					}()
				}
			}
			stew_table_lk.Unlock()
		}
	}()
	kilog.Debug("before main goroutine")
	go func() {
		for {
			thing, err := listener.Accept()
			if err != nil {
				continue
			}
			go func() {
				// Obtain the stew id
				idbuf := make([]byte, 16)
				_, err := io.ReadFull(thing, idbuf)
				if err != nil {
					thing.Close()
					kilog.Warning("Client didn't send the entire stew ID before dying")
					return
				}
				kilog.Debug("Obtained new SC on server with id=%x", idbuf)
				top := binary.BigEndian.Uint64(idbuf[0:8])
				bot := binary.BigEndian.Uint64(idbuf[8:16])
				id := stid{top, bot}
				stew_table_lk.Lock()
				if stew_table[id] == nil {
					stew_table[id] = make_stew_ctx()
					go stew_table[id].run_stew(true)
				}
				xaxa := stew_table[id]
				stew_table_lk.Unlock()
				xaxa.llctx.AttachSC(thing, true)
			}()
		}
	}()
	kilog.Debug("before return")
	return listener.Addr().String()
}

func single_stew_echo_server() {
	listener, err := net.Listen("tcp", "127.0.0.1:20999")
	if err != nil {
		panic(err.Error())
	}
	ctx := make_sc_ctx()

	go func() {
		for {
			ctx.write_ch <- <-ctx.ordered_ch
		}
	}()

	for {
		thing, err := listener.Accept()
		if err != nil {
			continue
		}
		go func() {
			ctx.AttachSC(thing, true)
		}()
	}
}

func single_stew_echo_client() {
	ctx := make_sc_ctx()
	sc1, err := net.Dial("tcp", "127.0.0.1:20999")
	if err != nil {
		panic(err.Error())
	}

	go ctx.AttachSC(sc1, false)
	sc2, err := net.Dial("tcp", "127.0.0.1:20999")
	if err != nil {
		panic(err.Error())
	}
	go ctx.AttachSC(sc2, false)
	kilog.Info("Two wires attached!")
	kilog.Info("Sending hello world...")
	ctx.write_ch <- sc_message{0x01, []byte(" ")}
	close(<-ctx.close_ch_ch)
	ctx.write_ch <- sc_message{0x00, []byte("hello")}
	sc3, err := net.Dial("tcp", "127.0.0.1:20999")
	if err != nil {
		panic(err.Error())
	}
	go ctx.AttachSC(sc3, false)
	ctx.write_ch <- sc_message{0x02, []byte("world")}
	close(<-ctx.close_ch_ch)
	fmt.Print(string((<-ctx.ordered_ch).payload))
	fmt.Print(string((<-ctx.ordered_ch).payload))
	fmt.Println(string((<-ctx.ordered_ch).payload))
}
