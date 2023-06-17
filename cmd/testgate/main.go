// Command gateway forwards the incoming connections to the server.
package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const addr = ":8081"

var (
	serverAddr = flag.String("server", "localhost:8082", "server address")
)

func main() {
	flag.Parse()

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	revproxy(l)
}

func revproxy(l net.Listener) {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   *serverAddr,
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.Host)
		proxy.ServeHTTP(w, r)
	})

	log.Printf("Gateway listening on %s, forwarding to %s", addr, *serverAddr)
	if err := http.Serve(l, nil); err != nil {
		log.Fatal(err)
	}
}
