package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rusq/vhoster"
	"github.com/rusq/vhoster/apiserver"
)

func TestClient_Add(t *testing.T) {
	// Create a test server that returns a JSON response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/vhost/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req apiserver.AddRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if req.HostPrefix != "test" {
			t.Errorf("unexpected host prefix: %s", req.HostPrefix)
		}
		if req.Target != "http://localhost:8080" {
			t.Errorf("unexpected target: %s", req.Target)
		}
		resp := apiserver.AddResponse{Hostname: "test.endless.lol"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer ts.Close()

	// Create a client with the test server URL
	client, err := NewClient(ts.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Call the Add method with test data
	hostname, err := client.Add("test", "http://localhost:8080")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if hostname != "test.endless.lol" {
		t.Errorf("unexpected hostname: %s", hostname)
	}
}

func TestClient_Remove(t *testing.T) {
	// Create a test server that returns a JSON response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/vhost/test.endless.lol" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Create a client with the test server URL
	client, err := NewClient(ts.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Call the Remove method with test data
	err = client.Remove("test.endless.lol")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
}

func TestClient_List(t *testing.T) {
	// Create a test server that returns a JSON response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/vhost/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := apiserver.ListResponse{
			Hosts: []vhoster.Host{
				{Name: "test1.endless.lol", URI: vhoster.Must(vhoster.Parse("http://localhost:8080"))},
				{Name: "test2.endless.lol", URI: vhoster.Must(vhoster.Parse("http://localhost:8081"))},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer ts.Close()

	// Create a client with the test server URL
	client, err := NewClient(ts.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Call the List method and check the response
	hosts, err := client.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(hosts) != 2 {
		t.Errorf("unexpected number of hosts: %d", len(hosts))
	}
	if hosts[0].Name != "test1.endless.lol" {
		t.Errorf("unexpected hostname: %s", hosts[0].Name)
	}
	if hosts[0].URI.String() != "http://localhost:8080" {
		t.Errorf("unexpected target: %s", hosts[0].URI)
	}
	if hosts[1].Name != "test2.endless.lol" {
		t.Errorf("unexpected hostname: %s", hosts[1].Name)
	}
	if hosts[1].URI.String() != "http://localhost:8081" {
		t.Errorf("unexpected target: %s", hosts[1].URI)
	}
}
