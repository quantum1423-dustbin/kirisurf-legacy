package kiss

import "testing"

func BenchmarkDH_1536_GenKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dh_gen_key(1536)
	}
}

func BenchmarkDH_1536_GenSecret(b *testing.B) {
	xaxa := dh_gen_key(1536)
	for i := 0; i < b.N; i++ {
		dh_gen_secret(xaxa.priv, xaxa.pub)
	}
}

func BenchmarkDH_2048_GenKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dh_gen_key(2048)
	}
}

func BenchmarkDH_2048_GenSecret(b *testing.B) {
	xaxa := dh_gen_key(2048)
	for i := 0; i < b.N; i++ {
		dh_gen_secret(xaxa.priv, xaxa.pub)
	}
}

func TestDH2048(t *testing.T) {
	goo := dh_gen_key(2048)
	doo := dh_gen_key(2048)
	xaxa := dh_gen_secret(goo.priv, doo.pub)
	gaga := dh_gen_secret(doo.priv, goo.pub)
	for i := 0; i < len(xaxa); i++ {
		if xaxa[i] != gaga[i] {
			t.FailNow()
		}
	}
}

func TestDH1536(t *testing.T) {
	goo := dh_gen_key(1536)
	doo := dh_gen_key(1536)
	xaxa := dh_gen_secret(goo.priv, doo.pub)
	gaga := dh_gen_secret(doo.priv, goo.pub)
	for i := 0; i < len(xaxa); i++ {
		if xaxa[i] != gaga[i] {
			t.FailNow()
		}
	}
}
