package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/rusq/vhoster"
)

const addr = ":8081"

func main() {
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	s, err := vhoster.Listen(addr)
	if err != nil {
		panic(err)
	}

	var sig = make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		s.Close()
		log.Printf("server stopped")
	}()

	uri1, _ := url.Parse("http://localhost:8082")
	uri2, _ := url.Parse("http://localhost:8082/elsewhere")
	s.Add("mm.localhost:8081", uri1)
	s.Add("localhost:8081", uri2)

	log.Printf("listening on %s", addr)
	s.Wait()
}
