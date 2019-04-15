package http

import (
	"net"
	nethttp "net/http"
	"strconv"
	"time"
)

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

// ListenOnAnyPort serve http on any port, return port by retPort if not nil
func ListenOnAnyPort(h nethttp.Handler, retPort *string) error {
	server := nethttp.Server{Handler: h}
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	server.Addr = ":" + strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	if retPort != nil {
		*retPort = server.Addr
	}
	return server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
}
