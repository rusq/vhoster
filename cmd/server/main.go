package main

import (
	"log"
	"net/http"
)

const addr = ":8082"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	http.HandleFunc("/elsewhere/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, Elsewhere!"))
	})
	log.Printf("Starting server on %s", addr)
	http.ListenAndServe(addr, nil)
}
