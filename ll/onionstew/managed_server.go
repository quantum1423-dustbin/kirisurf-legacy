package onionstew

import "io"

type ManagedServer struct {
	scchan chan io.ReadWriteCloser
}
