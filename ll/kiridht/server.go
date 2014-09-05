package kiridht

import (
	"bufio"
	"encoding/binary"
	"io"
	"strings"

	"github.com/KirisurfProject/kilog"
)

func HandleServer(conn io.ReadWriteCloser) {
	defer conn.Close()
	r, w := bufio.NewReader(conn), bufio.NewWriter(conn)
	for {
		l, e := r.ReadString('\n')
		if e != nil {
			return
		}
		l = l[0 : len(l)-1]
		line := strings.Split(l, " ")
		switch line[0] {
		case "GET":
			key := []byte(line[1])
			if len(key) != 8 {
				return
			}
			keyy := binary.LittleEndian.Uint64(key)
			if cacheIdx(keyy) == nil {
				kilog.Warning("DHT: No propagation implemented yet")
				return
			}
			_, err := w.Write(cacheIdx(keyy))
			if err != nil {
				return
			}
			_, err = w.WriteString("!end\n")
			if err != nil {
				return
			}
			w.Flush()
		case "SET":
			key := []byte(line[1])
			if len(key) != 8 {
				return
			}
			keyy := binary.LittleEndian.Uint64(key)
			val := make([]byte, 0)
			for {
				line, err := r.ReadString('\n')
				if err != nil {
					return
				}
				if line == "!end\n" {
					break
				}
				val = append(val, []byte(line)...)
				if len(val) >= 1024*1024*4 {
					return
				}
			}
			cacheAdd(keyy, val)
			w.WriteString("DUN\n")
			w.Flush()
		default:
			return
		}
	}
}
