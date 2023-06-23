package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/rusq/osenv/v2"
)

var (
	host = flag.String("host", osenv.Value("TESTSERVER_HOST", "localhost"), "host to listen on")
	port = flag.String("port", osenv.Value("TESTSERVER_PORT", "8082"), "port to listen on")
)

func main() {
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	http.HandleFunc("/elsewhere/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, Elsewhere!"))
	})
	addr := *host + ":" + *port
	log.Printf("Starting server on %s", addr)
	http.ListenAndServe(addr, nil)
}
