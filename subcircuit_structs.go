package main

const (
	SC_EXTEND    = iota
	SC_TERMINATE = iota
)

type sc_message struct {
	msg_type int
	msg_arg  string
}
