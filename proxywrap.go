package vhoster

import (
	"context"
	"net"
	"net/http"
	"sync"
)

type proxyWrapper struct {
	vhost Host
	l     net.Listener
	srv   *http.Server
	wg    *sync.WaitGroup // reference to the parent waitgroup
}

// Close closes all open handles and connections.
func (pw proxyWrapper) Close() error {
	pw.srv.Shutdown(context.Background())
	pw.l.Close()
	pw.wg.Done()
	return nil
}
