package vhoster

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"sync"
)

type proxyWrapper struct {
	VHost string
	URI   *url.URL
	l     net.Listener
	srv   *http.Server
	wg    *sync.WaitGroup // reference to the parent waitgroup
}

func (pw proxyWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pw.srv.Handler.ServeHTTP(w, r)
}

func (pw proxyWrapper) Close() error {
	pw.srv.Shutdown(context.Background())
	pw.l.Close()
	pw.wg.Done()
	return nil
}
