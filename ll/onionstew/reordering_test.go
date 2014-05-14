package onionstew

import "testing"

// TestReorderingBasic tests reorder_messages functionality.
func TestReordering(t *testing.T) {
	in := make(chan sc_message)
	out := make(chan sc_message)
	go func() {
		for i := 1000; i >= 0; i = i - 2 {
			thing := sc_message{uint64(i), []byte("")}
			in <- thing
		}
	}()
	go func() {
		for i := 1; i <= 2999; i = i + 2 {
			thing := sc_message{uint64(i), []byte("")}
			in <- thing
		}
		close(in)
	}()
	go reorder_messages(in, out)
	last := 0
	for {
		bla, ok := <-out
		if !ok {
			return
		}
		if int(bla.seqnum) < last {
			t.FailNow()
		}
		last = int(bla.seqnum)
	}
}

// TestReorderingError tests error when reorder buffer is blown.
func TestReorderingError(t *testing.T) {
	in := make(chan sc_message)
	out := make(chan sc_message)
	go func() {
		_, ok := <-out
		if ok {
			t.FailNow()
		}
	}()
	go func() {
		for i := 1; i < 1000000; i++ {
			ba := sc_message{uint64(i), []byte("")}
			in <- ba
		}
	}()
	reorder_messages(in, out)
}
