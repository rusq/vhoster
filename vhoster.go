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

// Server is a virtual host reverse proxy server.
type Server struct {
	// Addr is the server address
	Addr string
	ln   net.Listener // main listener
	vhm  *vhost.HTTPMuxer
	done chan struct{}

	mu  sync.Mutex
	pws map[string]proxyWrapper // a map of registered listeners
	wg  *sync.WaitGroup         // a waitgroup for running servers
}

// Listen initialises the server and starts listening on the given address.
func Listen(addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	vhm, err := vhost.NewHTTPMuxer(ln, 100*time.Millisecond)
	if err != nil {
		return nil, err
	}
	done := make(chan struct{})
	go errorhandler(vhm, done)
	return &Server{
		Addr: addr,
		ln:   ln,
		vhm:  vhm,
		done: done,
		pws:  make(map[string]proxyWrapper, 1),
		wg:   new(sync.WaitGroup),
	}, nil
}

func (s *Server) Close() error {
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
func (s *Server) Add(vhost string, uri *url.URL) error {
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
	s.wg.Add(1)
	pw := proxyWrapper{
		l:     ml,
		srv:   &srv,
		wg:    s.wg,
		VHost: vhost,
		URI:   uri,
	}
	s.pws[vhost] = pw

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

// Remove removes the virtual host from the server.
func (s *Server) Remove(vhost string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, ok := s.pws[vhost]
	if !ok {
		return errors.New("vhost not found")
	}
	delete(s.pws, vhost)
	return l.Close()
}

// Wait blocks until the server is closed.
func (s *Server) Wait() {
	<-s.done
}

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
			handleError(conn, errors.New("bad request"))
		case vhost.NotFound:
			log.Printf("got a connection for an unknown vhost: %s", err)
			handleError(conn, errors.New("vhost not found"))
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
				handleError(conn, errors.New("server error"))
			}
		}
		if conn != nil {
			conn.Close()
		}
	}
}

func handleError(conn net.Conn, err error) {
	// Create a new HTTP response object.
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError, // TODO
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

type Host struct {
	Name string
	URI  *url.URL
}

func (s *Server) List() []Host {
	s.mu.Lock()
	defer s.mu.Unlock()

	var vhosts []Host
	for vhost, pw := range s.pws {
		vhosts = append(vhosts, Host{
			Name: vhost,
			URI:  pw.URI,
		})
	}
	return vhosts
}
