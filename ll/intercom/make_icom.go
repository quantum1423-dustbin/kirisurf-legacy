package intercom

import (
	"encoding/binary"
	"io"
	"sync"

	"github.com/KirisurfProject/kilog"
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

func (xaxa *icom_msg) WriteTo(writer io.Writer) error {
	scratch := make([]byte, len(xaxa.body)+5)
	scratch[0] = byte(xaxa.flag)
	binary.LittleEndian.PutUint16(scratch[1:3], uint16(xaxa.connid))
	binary.LittleEndian.PutUint16(scratch[3:5], uint16(len(xaxa.body)))
	copy(scratch[5:], xaxa.body)
	_, err := writer.Write(scratch)
	return err
}

func (xaxa *icom_msg) ReadFrom(reader io.Reader) error {
	mdat := make([]byte, 3)
	_, err := io.ReadFull(reader, mdat)
	if err != nil {
		return err
	}
	xaxa.flag = int(mdat[0])
	xaxa.connid = int(binary.LittleEndian.Uint16(mdat[1:3]))
	_, err = io.ReadFull(reader, mdat[0:2])
	if err != nil {
		return err
	}
	length := binary.LittleEndian.Uint16(mdat[0:2])
	if err != nil {
		return err
	}
	body := make([]byte, int(length))
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return err
	}
	xaxa.body = body
	return nil
}

const (
	icom_ignore = 0x00

	icom_open  = 0x10
	icom_close = 0x11
	icom_data  = 0x12
	icom_more  = 0x13
)

func make_icom_ctx(underlying io.ReadWriteCloser, is_server bool, do_junk bool) *icom_ctx {
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
			kilog.Debug("Killswitching!")
			ctx.underlying.Close()
			ctx.is_dead = true
			close(killswitch)
			close(ctx.our_srv.vs_ch)
		})
	}

	// Run the main thing
	go run_icom_ctx(ctx, KILL, is_server, do_junk)

	return ctx
}
