// common.go
package kiss

//import "code.google.com/p/go-sqlite/go1/sqlite3"
import (
	"fmt"
	"kirisurf/ll/buf"
)
import "net"
import "time"
import "errors"
import "runtime"
import "path/filepath"
import "os"
import "io"
import "encoding/binary"

var LOG_ERROR int = 0
var LOG_NOTICE int = 1
var LOG_INFO int = 2
var LOG_DEBUG int = 3

var _log_lock chan bool = make(chan bool, 1)

func LOG(x int, format string, a ...interface{}) {
	return
	_log_lock <- true
	switch {
	case x == LOG_DEBUG:
		fmt.Print("debug at")
	case x == LOG_NOTICE:
		fmt.Print("notice at")
	case x == LOG_ERROR:
		fmt.Print("error at")
	case x == LOG_INFO:
		fmt.Print("info at")
	}
	_, file, line, _ := runtime.Caller(1)
	toret := fmt.Sprintf(format, a...)
	fmt.Printf(" %s:%d : ", filepath.Base(file), line)
	fmt.Println(toret)
	<-_log_lock
}

func copy_conns(xconn io.ReadCloser, yconn io.WriteCloser) {
	buff := buf.Alloc()
	defer func() {
		xconn.Close()
		yconn.Close()
		buf.Free(buff)
	}()
	for {
		//runtime.Gosched()
		l, e := xconn.Read(buff)
		if e != nil {
			return
		}
		l, e = yconn.Write(buff[:l])
		if e != nil {
			return
		}
	}
}

func obfs_accept(listener net.Listener) (net.Conn, error) {
	c, err := listener.Accept()
	if err != nil {
		return nil, err
	}
	conn := net.Conn(nil)
	donechan := make(chan error)
	eee := error(nil)
	go func() {
		conn, err = Kiriobfs_handshake_server(c)
		eee = err
		donechan <- err
	}()
	select {
	case <-donechan:
		return conn, eee
	case <-time.After(time.Second * time.Duration(10)):
		return nil, errors.New("timed out")
	}
}

func server_with_dispatch(addr string, handler func(net.Conn)) net.Listener {
	listener, err := net.Listen("tcp", addr)
	check_fatal(err)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				LOG(LOG_ERROR, "server at %s encountered error while accepting: %s",
					addr,
					err.Error())
				listener.Close()
				return
			}
			go func() {
				defer func() {
					conn.Close()
				}()
				handler(conn)
			}()
		}
	}()
	return listener
}

func open_file(path string) *os.File {
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 666)
	if err != nil {
		panic(err.Error())
	}
	return fd
}

func check_serious(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "**** KURWA!!! ****\n")
		fmt.Fprintf(os.Stderr, "Error encountered was of SERIOUS, cannot into continue. Goroutine of killed, deferred resources of cleaned.\n")
		fmt.Fprintf(os.Stderr, "Error details: %s\n", err.Error())
		runtime.Goexit()
	}
}

func check_fatal(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "**** KURWA!!! ****\n")
		fmt.Fprintf(os.Stderr, "Error encountered was of FATAL. Is of wrong. Gib correct config pl0x.\n")
		panic(err.Error())
	}
}

func SPANIC(blerg string) {
	check_serious(errors.New(blerg))
}

func big_endian_uint16(num int) []byte {
	toret := make([]byte, 2)
	binary.BigEndian.PutUint16(toret, uint16(num))
	return toret
}

func little_endian_uint64(num uint64) []byte {
	toret := make([]byte, 32)
	binary.LittleEndian.PutUint64(toret, uint64(num))
	return toret
}

func FASSERT(cond bool) {
	_, file, line, _ := runtime.Caller(1)
	if !cond {
		fmt.Fprintf(os.Stderr, "**** KURWA!!! ****\n")
		fmt.Fprintf(os.Stderr, "Error encountered was of ASSERTION. Kirisurf cannot into continue. Gib bugfix fast pl0x.\n")
		fmt.Fprintf(os.Stderr, "Bad FASSERT called in %s:%d\n", filepath.Base(file), line)
		panic("bad assert")
	}
}

func SASSERT(cond bool) {
	_, file, line, _ := runtime.Caller(1)
	if !cond {
		SPANIC(fmt.Sprintf("bad sassert at %s:%d", filepath.Base(file), line))
	}
}
