package vhoster

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/go-vhost"
)

// Gateway is a virtual host reverse proxy server.  Zero value is not usable.
type Gateway struct {
	ln   net.Listener // main listener
	vhm  *vhost.HTTPMuxer
	done chan struct{}

	mu  sync.Mutex
	pws map[string]proxyWrapper // a map of registered listeners
	wg  *sync.WaitGroup         // a waitgroup for running servers
}

// Host is a single Virtual Host.
type Host struct {
	// Name is the name of the Virtual Host.
	Name string
	// URI is the URI of the target HTTP server.
	URI *url.URL
}

// Option is a functional option for the server.
type Option func(*options)

type options struct {
	timeout time.Duration
	hosts   []Host
}

// WithTimeout sets the connection timeout to the virtual hosts.
func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.timeout = d
		}
	}
}

func WithHosts(hs []Host) Option {
	return func(o *options) {
		o.hosts = hs
	}
}

// Listen initialises the server and starts listening on the given address.
func Listen(addr string, opts ...Option) (*Gateway, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	o := &options{
		timeout: 100 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(o)
	}

	vhm, err := vhost.NewHTTPMuxer(ln, o.timeout)
	if err != nil {
		return nil, err
	}
	done := make(chan struct{})
	g := &Gateway{
		ln:   ln,
		vhm:  vhm,
		done: done,
		pws:  make(map[string]proxyWrapper, 1),
		wg:   new(sync.WaitGroup),
	}

	// preconfigured hosts
	for _, h := range o.hosts {
		if err := g.Add(h.Name, h.URI); err != nil {
			return nil, err
		}
	}

	go errorhandler(vhm, done)
	return g, nil
}

func (s *Gateway) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer close(s.done)

	// closing listeners
	for vhost, l := range s.pws {
		delete(s.pws, vhost)
		l.Close()
	}
	s.wg.Wait() // waiting for servers to shut down
	s.vhm.Close()
	return s.ln.Close()
}

// Add adds the virtual host to the server.
func (s *Gateway) Add(vhost string, uri *url.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lg := log.New(log.Default().Writer(), vhost+": ", log.Default().Flags())

	lg.Printf("setting up proxy for %s to %s", vhost, uri)
	ml, err := s.vhm.Listen(vhost)
	if err != nil {
		return err
	}
	srv := http.Server{
		Handler: httputil.NewSingleHostReverseProxy(uri),
	}
	pw := proxyWrapper{
		l:   ml,
		srv: &srv,
		wg:  s.wg,
		vhost: Host{
			Name: vhost,
			URI:  uri,
		},
	}
	s.pws[vhost] = pw

	s.wg.Add(1)
	go func() {
		if err := srv.Serve(ml); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			lg.Printf("error: %v", err)
		}
	}()
	return nil
}

var ErrNotFound = errors.New("vhost not found")

// Remove removes the virtual host from the server.
func (s *Gateway) Remove(vhost string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, ok := s.pws[vhost]
	if !ok {
		return ErrNotFound
	}
	delete(s.pws, vhost)
	return l.Close()
}

// Wait blocks until the server is closed.
func (s *Gateway) Wait() {
	<-s.done
}

// errorhandler loops over the errors returned by the vhost manager
// and handles them, if necessary.  It exists when done channel is
// closed.
func errorhandler(vm *vhost.HTTPMuxer, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
		}
		conn, err := vm.NextError()
		switch err.(type) {
		case vhost.BadRequest:
			log.Print("got a bad request!")
			handleError(conn, http.StatusBadRequest, errors.New("bad request"))
		case vhost.NotFound:
			log.Printf("got a connection for an unknown vhost: %s", err)
			handleError(conn, http.StatusNotFound, ErrNotFound)
		case vhost.Closed:
			log.Printf("closed conn: %s", err)
		default:
			var opErr *net.OpError
			if errors.As(err, &opErr) {
				if strings.EqualFold(opErr.Err.Error(), "use of closed network connection") {
					continue
				}
			}
			log.Printf("generic error (%[1]T): %[1]s,", err)
			if conn != nil {
				handleError(conn, http.StatusInternalServerError, errors.New("server error"))
			}
		}
		if conn != nil {
			conn.Close()
		}
	}
}

func handleError(conn net.Conn, code int, err error) {
	// Create a new HTTP response object.
	if code == 0 {
		code = http.StatusInternalServerError
	}
	resp := &http.Response{
		StatusCode: code,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(err.Error())),
	}

	// Write the response to the connection.
	if err := resp.Write(conn); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (s *Gateway) List() []Host {
	s.mu.Lock()
	defer s.mu.Unlock()

	var vhosts []Host
	for _, pw := range s.pws {
		vhosts = append(vhosts, pw.vhost)
	}
	return vhosts
}
