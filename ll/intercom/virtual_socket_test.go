package intercom

import (
	"io"
	"testing"
)

func BenchmarkBP(b *testing.B) {
	xaxa := NewBufferedPipe()
	go func() {
		defer xaxa.Close()
		lel := make([]byte, 1024)
		for {
			_, err := xaxa.Write(lel)
			if err != nil {
				return
			}
		}
	}()
	defer xaxa.Close()
	b.ResetTimer()
	lel := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		io.ReadFull(xaxa, lel)
	}
}
