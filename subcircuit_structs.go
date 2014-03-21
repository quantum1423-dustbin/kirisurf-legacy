package main

const (
	SC_EXTEND    = iota
	SC_TERMINATE = iota
)

type sc_message struct {
	Msg_type int
	Msg_arg  string
}
