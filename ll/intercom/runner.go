package intercom

import (
	"io"
	"sync"
)

type icom_ctx struct {
	is_dead    bool
	underlying io.ReadWriteCloser
	rlock      sync.Mutex
	wlock      sync.Mutex
	our_srv    *VirtualServer
	write_ch   chan icom_msg
	killswitch chan bool
}

type icom_msg struct {
	flag   int
	connid int
	body   []byte
}

const (
	icom_ignore = 0x00

	icom_open  = 0x10
	icom_close = 0x11
	icom_data  = 0x12
	icom_more  = 0x13
)

func make_icom_ctx(underlying io.ReadWriteCloser, is_server bool) *icom_ctx {
	ctx := new(icom_ctx)
	ctx.is_dead = false
	ctx.underlying = underlying
	ctx.our_srv = VSListen()
	ctx.write_ch = make(chan icom_msg)

	// Killswitch is closed when the entire ctx should be abandoned.
	killswitch := make(chan bool)
	ctx.killswitch = killswitch
	var _ks_exec sync.Once
	KILL := func() {
		_ks_exec.Do(func() {
			ctx.underlying.Close()
			ctx.is_dead = true
			close(killswitch)
			close(ctx.our_srv.vs_ch)
		})
	}

	// Run the main thing
	go run_icom_ctx(ctx, KILL, is_server)

	return ctx
}
