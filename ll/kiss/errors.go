package kiss

import "errors"

var ErrPacketTooShort = errors.New("Packet less than 44 bytes! Kurwa!")

var ErrMacNoMatch = errors.New("MAC error! Dieee!")
