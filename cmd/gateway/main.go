package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/rusq/vhoster"
)

var (
	host    = flag.String("gateway-host", "localhost", "host to listen on")
	port    = flag.String("gateway-port", "8081", "port to listen on")
	pubAddr = flag.String("pub", "localhost:8081", "server public `hostname`, it is used as a suffix for all vhosts")

	apiaddr = flag.String("api", "localhost:8083", "address of this api server that controls the gateway")
)

func main() {
	flag.Parse()

	s, err := vhoster.Listen(*host + ":" + *port)
	if err != nil {
		log.Fatal(err)
	}
	go s.Wait()

	gateway := &gateway{
		addr: *pubAddr,
		vs:   s,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/vhost/", Only(gateway.handleVhost, http.MethodPost, http.MethodDelete, http.MethodGet))
	mux.HandleFunc("/random/", Only(gateway.handleRandom, http.MethodPost))
	mux.HandleFunc("/health/", Only(gateway.handleHealth, http.MethodGet))

	log.Printf("listening on %s", *apiaddr)
	log.Fatal(http.ListenAndServe(*apiaddr, mux))
}

type vhostmgr interface {
	Add(string, *url.URL) error
	Remove(string) error
	List() []vhoster.Host
}

type gateway struct {
	addr string
	vs   vhostmgr
}

type AddRequest struct {
	HostPrefix string `json:"host_prefix,omitempty"`
	Target     string `json:"target,omitempty"`
}

type AddResponse struct {
	Hostname string `json:"hostname,omitempty"`
}

func Only(h http.HandlerFunc, methods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, m := range methods {
			if r.Method == m {
				h(w, r)
				return
			}
		}
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (g *gateway) handleVhost(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// create
		g.handleAdd(w, r)
	case http.MethodDelete:
		// delete
		g.handleRemove(w, r)
	case http.MethodGet:
		// list
		g.handleList(w, r)
	}
}

func (g *gateway) handleAdd(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Print("error decoding body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	g.processAdd(w, r, &req)
}

func (g *gateway) processAdd(w http.ResponseWriter, r *http.Request, req *AddRequest) {
	if req.Target == "" {
		log.Print("empty target")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	uri, err := url.Parse(req.Target)
	if err != nil {
		log.Print("error parsing the target hostname:", err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	vhost := req.HostPrefix + "." + g.addr
	if _, err := url.Parse(vhost); err != nil {
		log.Print("error parsing the resulting hostname:", err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	if err := g.vs.Add(vhost, uri); err != nil {
		log.Print("error adding host:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AddResponse{Hostname: vhost})
}

type vhost struct {
	Host string `json:"host,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type ListResponse struct {
	Hosts []vhost
}

func (g *gateway) handleList(w http.ResponseWriter, r *http.Request) {
	hosts := g.vs.List()
	w.Header().Set("Content-Type", "application/json")
	var resp ListResponse
	for _, h := range hosts {
		resp.Hosts = append(resp.Hosts, vhost{Host: h.Name, URI: h.URI.String()})
	}
	json.NewEncoder(w).Encode(resp)
}

type RandomResponse struct {
	AddResponse
}

type RandomRequest struct {
	Target string `json:"target,omitempty"`
}

// handleRandom creates a random hostname and adds it to the gateway
// it returns the hostname
func (g *gateway) handleRandom(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// generate a random uuid without using external libs
	baVhost := [16]byte{}
	for i := 0; i < 16; i++ {
		baVhost[i] = byte(rand.Intn(256))
	}
	sVhost := hex.EncodeToString(baVhost[:])
	var req RandomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Print("error decoding body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	g.processAdd(w, r, &AddRequest{HostPrefix: sVhost, Target: req.Target})
}

func (g *gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (g *gateway) handleRemove(w http.ResponseWriter, r *http.Request) {
	// remove
	vhost := r.URL.Path[len("/vhost/"):]
	if vhost == "" {
		log.Print("got empty vhost")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := g.vs.Remove(vhost + "." + g.addr); err != nil {
		log.Print("error removing host:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
