package main

import (
	"io"

	"github.com/KirisurfProject/kilog"
)

const (
	SC_EXTEND    = iota
	SC_TERMINATE = iota
)

type sc_message struct {
	Msg_type int
	Msg_arg  string
}

func read_sc_message(thing io.Reader) (sc_message, error) {
	var toret sc_message
	mslen := make([]byte, 1)
	_, err := io.ReadFull(thing, mslen)
	kilog.Debug("read mslen=%d", mslen[0])
	if err != nil {
		return toret, err
	}
	arg := make([]byte, mslen[0])
	_, err = io.ReadFull(thing, arg)
	kilog.Debug("read remaining")
	if err != nil {
		return toret, err
	}
	toret.Msg_type = int(arg[0])
	toret.Msg_arg = string(arg[1:])
	return toret, nil
}

func write_sc_message(msg sc_message, thing io.Writer) error {
	kilog.Debug("write_sc_message(%x)", msg)
	tosend := make([]byte, len([]byte(msg.Msg_arg))+2)
	tosend[0] = byte(len(tosend) - 1)
	tosend[1] = byte(msg.Msg_type)
	copy(tosend[2:], []byte(msg.Msg_arg))
	_, err := thing.Write(tosend)
	return err
}
