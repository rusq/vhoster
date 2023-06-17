package main

import (
	"log"
	"net"
	"time"

	"github.com/inconshreveable/go-vhost"
)

var virtualHosts []virtualHost = []virtualHost{
	&A{name: "abc.ee.com.se:55544"},
}

func main() {
	l, err := net.Listen("tcp", ":55544")
	if err != nil {
		panic(err)
	}
	x(l)
}

func x(l net.Listener) {
	mux, err := vhost.NewHTTPMuxer(l, 100*time.Millisecond)
	if err != nil {
		panic(err)
	}
	// listen for connections to different domains
	for _, v := range virtualHosts {
		vhost := v

		// vhost.Name is a virtual hostname like "foo.example.com"
		muxListener, _ := mux.Listen(vhost.Name())

		go func(vh virtualHost, ml net.Listener) {
			for {
				conn, _ := ml.Accept()
				go vh.Handle(conn)
			}
		}(vhost, muxListener)
	}

	for {
		conn, err := mux.NextError()

		switch err.(type) {
		case vhost.BadRequest:
			log.Printf("got a bad request!")
			conn.Write([]byte("bad request"))
		case vhost.NotFound:
			log.Printf("got a connection for an unknown vhost: %s", err)
			conn.Write([]byte("vhost not found"))
		case vhost.Closed:
			log.Printf("closed conn: %s", err)
		default:
			if conn != nil {
				conn.Write([]byte("server error"))
			}
		}

		if conn != nil {
			conn.Close()
		}
	}
}
