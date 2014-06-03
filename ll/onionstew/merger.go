package onionstew

import "github.com/KirisurfProject/kilog"

func reorder_messages(input, output chan sc_message) {
	defer close(output)
	ptr := uint64(0)
	buffer := make(map[uint64]sc_message)
	for {
		thing, ok := <-input
		if !ok {
			return
		}
		buffer[thing.seqnum] = thing
		if len(buffer) > 4096 {
			kilog.Critical("reorder_messages buffer blown")
			return
		}
		for {
			thing, ok := buffer[ptr]
			if !ok {
				break
			}
			delete(buffer, ptr)
			output <- thing
			ptr++
		}
	}
}
