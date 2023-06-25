package apiserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/rusq/vhoster"
)

func Run(vg HostManager, apiAddr, pubAddr string) error {
	gw := &gateway{
		addr: pubAddr,
		vg:   vg,
	}
	return http.ListenAndServe(apiAddr, gw.handler())
}

func (g *gateway) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/vhost/", Only(g.handleVhost, http.MethodPost, http.MethodDelete, http.MethodGet, http.MethodPatch))
	mux.HandleFunc("/random/", Only(g.handleRandom, http.MethodPost))
	mux.HandleFunc("/health/", Only(g.handleHealth, http.MethodGet))
	return mux
}

//go:generate mockgen -destination=../mocks/mock_hostmanager.go -package=mocks github.com/rusq/vhoster/apiserver HostManager
type HostManager interface {
	Add(string, *url.URL) error
	Remove(string) error
	Replace(string, *url.URL) error
	List() []vhoster.Host
}

type gateway struct {
	addr string
	vg   HostManager
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
	case http.MethodPatch:
		// update
		g.handleReplace(w, r)
	case http.MethodDelete:
		// delete
		g.handleRemove(w, r)
	case http.MethodGet:
		// list
		g.handleList(w, r)
	}
}

func httpErr(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Error(w, http.StatusText(code), code)
}

func (g *gateway) handleAdd(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Print("error decoding body:", err)
		httpErr(w, http.StatusBadRequest)
		return
	}
	g.process(w, r, &req, g.vg.Add)
}

type ReplaceRequest AddRequest

type ReplaceResponse AddResponse

func (g *gateway) handleReplace(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req ReplaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Print("error decoding body:", err)
		httpErr(w, http.StatusBadRequest)
		return
	}
	g.process(w, r, (*AddRequest)(&req), g.vg.Replace)
}

func (g *gateway) process(w http.ResponseWriter, r *http.Request, req *AddRequest, fn func(string, *url.URL) error) {
	if req.Target == "" {
		log.Print("missing target")
		http.Error(w, "400 missing target", http.StatusBadRequest)
		return
	}
	uri, err := url.Parse(req.Target)
	if err != nil {
		log.Print("error parsing the target hostname:", err)
		http.Error(w, "400 invalid target", http.StatusBadRequest)
		return
	}

	vhost := g.withDomain(req.HostPrefix)
	if _, err := url.Parse(vhost); err != nil {
		log.Printf("error parsing the resulting hostname %q: %s", vhost, err)
		http.Error(w, "400 invalid host prefix", http.StatusBadRequest)
		return
	}
	if err := fn(vhost, uri); err != nil {
		log.Printf("error adding host %q: %s", vhost, err)
		if errors.Is(err, vhoster.ErrAlreadyExists) {
			http.Error(w, "409 host already exists", http.StatusConflict)
			return
		}
		httpErr(w, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(AddResponse{Hostname: vhost}); err != nil {
		log.Printf("error encoding response for vhost %q: %s", vhost, err)
		httpErr(w, http.StatusInternalServerError)
		return
	}
}

func (g *gateway) withDomain(hostprefix string) string {
	return hostprefix + "." + g.addr
}

func (g *gateway) handleList(w http.ResponseWriter, r *http.Request) {
	vHost := vhostName(r)
	hosts := g.vg.List()
	if vHost == "" {
		g.listHosts(w, hosts)
		return
	}
	fullName := g.withDomain(vHost)
	for _, h := range hosts {
		if h.Name == fullName || h.Name == vHost {
			g.listHosts(w, []vhoster.Host{h})
			return
		}
	}
	http.NotFound(w, r)
}

// ListResponse is a response for the list request.
type ListResponse struct {
	Hosts []vhoster.Host `json:"hosts,omitempty"`
}

// listHosts encodes the list of hosts to the response.
func (g *gateway) listHosts(w http.ResponseWriter, hosts []vhoster.Host) {
	w.Header().Set("Content-Type", "application/json")
	var resp = ListResponse{
		Hosts: hosts,
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
	h := randString(16)
	var req RandomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Print("error decoding body:", err)
		http.Error(w, "error decoding body", http.StatusBadRequest)
		return
	}
	g.process(w, r, &AddRequest{HostPrefix: h, Target: req.Target}, g.vg.Add)
}

var randString = func(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(256))
	}

	return hex.EncodeToString(b)
}

func (g *gateway) handleHealth(w http.ResponseWriter, _ *http.Request) {
	httpErr(w, http.StatusOK)
}

func vhostName(r *http.Request) string {
	return r.URL.Path[len("/vhost/"):]
}

func (g *gateway) handleRemove(w http.ResponseWriter, r *http.Request) {
	// remove
	vhost := vhostName(r)
	if vhost == "" {
		log.Print("got empty vhost")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := g.vg.Remove(vhost); err != nil {
		if errors.Is(err, vhoster.ErrNotFound) {
			// attempt to remove by prefix
			log.Printf("host %q does not exist, assuming prefix", vhost)
			if err := g.vg.Remove(vhost + "." + g.addr); err != nil {
				log.Print("error removing host:", err)
			}
		} else {
			log.Print("error removing host:", err)
			httpErr(w, http.StatusInternalServerError)
			return
		}
		http.Error(w, "host does not exist", http.StatusNotFound)
		return
	}
	httpErr(w, http.StatusOK)
}
