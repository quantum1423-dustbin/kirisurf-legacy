package intercom

import (
	"fmt"
	"io"
	"sync"
)

func icom_tunnel(ctx *icom_ctx, KILL func(), conn io.ReadWriteCloser,
	connid int, reader chan icom_msg, do_junk bool) {

	PAUSELIM := 128
	if do_junk {
		PAUSELIM = 512
	}
	if !do_junk {
		fmt.Println(PAUSELIM)
	}

	local_close := make(chan bool)
	var _thing sync.Once
	local_kill := func() {
		_thing.Do(func() {
			close(local_close)
		})
	}

	// Kill local when returns
	defer local_kill()

	// Kill local when global dies
	go func() {
		select {
		case <-ctx.killswitch:
			local_kill()
		case <-local_close:
		}
	}()

	// Kill connection when local dies
	go func() {
		<-local_close
		conn.Close()
	}()

	// Semaphore for send flow control
	fctl := make(chan bool, PAUSELIM)
	for i := 0; i < PAUSELIM; i++ {
		select {
		case fctl <- true:
		default:
		}
	}
	// De-encapsulate
	go func() {
		defer local_kill()
		i := uint64(0)
		for {
			select {
			case <-local_close:
				return
			case pkt := <-reader:
				if pkt.flag == icom_close {
					return
				} else if pkt.flag == icom_data {
					i++
					// Is of data. Into puttings.
					_, err := conn.Write(pkt.body)
					if err != nil {
						return
					}
					if i%uint64(PAUSELIM) == 0 {
						go func() {
							select {
							case ctx.write_ch <- icom_msg{icom_more, connid,
								make([]byte, 0)}:
								fmt.Println("Sent icom_more")
							case <-ctx.killswitch:
							}
						}()
					}
				} else if pkt.flag == icom_more {
					fmt.Println("Got icom_more")
					for i := 0; i < PAUSELIM; i++ {
						select {
						case fctl <- true:
						default:
						}
					}
				}
			}
		}
	}()

	// Encapsulate
	func() {
		defer local_kill()
		buff := make([]byte, 2048)
		for {
			select {
			case <-local_close:
				return
			default:
				n, err := conn.Read(buff)
				if err != nil {
					select {
					case ctx.write_ch <- icom_msg{icom_close, connid, make([]byte, 0)}:
					case <-local_close:
						return
					}
					return
				}
				xaxa := make([]byte, n)
				copy(xaxa, buff)
				if !do_junk {
					fmt.Println("!do_junk, fctl=%d", len(fctl))
				}
				select {
				case <-fctl:
				case <-local_close:
					return
				}
				select {
				case ctx.write_ch <- icom_msg{icom_data, connid, xaxa}:
				case <-local_close:
					return
				}
			}
		}
	}()
}
