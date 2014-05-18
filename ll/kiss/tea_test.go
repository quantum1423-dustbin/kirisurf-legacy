package kiss

import "testing"

/*
func TestTEA(t *testing.T) {
	val := make([]byte, 8)
	key := make([]byte, 16)
	tea_encrypt(val, key)
	fmt.Printf("%x\n", val)
	val = make([]byte, 8)
	key = make([]byte, 16)
	tea_encrypt(val, key)
	fmt.Printf("%x\n", val)
}*/

func BenchmarkTEA(b *testing.B) {
	val := make([]byte, 8)
	key := make([]byte, 16)
	xaxa := make_tea(key)
	for i := 0; i < b.N; i++ {
		for i := 0; i < 1024/8; i++ {
			xaxa.Encrypt(val, val)
		}
	}
}
