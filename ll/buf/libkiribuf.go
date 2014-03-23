// libkiribuf project libkiribuf.go
package buf

var ___global_buffer_chan = make(chan []byte, 50)
var BSIZE = 16384

func Alloc() []byte {
	var b []byte
	select {
	case b = <-___global_buffer_chan:
	default:
		b = make([]byte, BSIZE)
	}
	return b
}

func Free(slc []byte) {
	if len(slc) != BSIZE {
		panic("Tried to free invalid buffer")
	}
	select {
	case ___global_buffer_chan <- slc:
	default:
	}
}
