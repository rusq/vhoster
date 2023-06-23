package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rusq/vhoster/apiserver"
)

var ErrNotFound = fmt.Errorf("not found")

var (
	epVhosts = &url.URL{Path: "/vhost/"}
	epRandom = &url.URL{Path: "/random/"}
)

func rVhostPath(name string) *url.URL {
	return &url.URL{Path: "/vhost/" + name}
}

type Client struct {
	base *url.URL
	cl   *http.Client
}

type Option func(*Client)

func WithHTTPClient(cl *http.Client) Option {
	return func(c *Client) {
		c.cl = cl
	}
}

func NewClient(baseURL string, opts ...Option) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{base: u, cl: http.DefaultClient}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

func (c *Client) Add(hostPrefix, target string) (string, error) {
	reqBody, err := json.Marshal(apiserver.AddRequest{
		HostPrefix: hostPrefix,
		Target:     target,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, c.base.ResolveReference(epVhosts).String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	var addResp apiserver.AddResponse
	if err := do(&addResp, c.cl, req); err != nil {
		return "", err
	}
	return addResp.Hostname, nil
}

// Random calls the /random endpoint of the server and returns a random
// hostname.
func (c *Client) Random() (string, error) {
	req, err := http.NewRequest(http.MethodGet, c.base.ResolveReference(epRandom).String(), nil)
	if err != nil {
		return "", err
	}
	var randomResp apiserver.RandomResponse
	if err := do(&randomResp, c.cl, req); err != nil {
		return "", err
	}
	return randomResp.Hostname, nil
}

func (c *Client) Remove(hostname string) error {
	rmHost := rVhostPath(hostname)
	req, err := http.NewRequest(http.MethodDelete, c.base.ResolveReference(rmHost).String(), nil)
	if err != nil {
		return err
	}
	resp, err := c.cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) List() ([]apiserver.ListHost, error) {
	req, err := http.NewRequest(http.MethodGet, c.base.ResolveReference(epVhosts).String(), nil)
	if err != nil {
		return nil, err
	}
	var listResp apiserver.ListResponse
	if err := do(&listResp, c.cl, req); err != nil {
		return nil, err
	}
	return listResp.Hosts, nil
}

func (c *Client) ListHost(prefix string) (*apiserver.ListHost, error) {
	listHost := rVhostPath(prefix)
	req, err := http.NewRequest(http.MethodGet, c.base.ResolveReference(listHost).String(), nil)
	if err != nil {
		return nil, err
	}
	var lr apiserver.ListResponse
	if err := do(&lr, c.cl, req); err != nil {
		return nil, err
	}
	if len(lr.Hosts) == 0 {
		return nil, ErrNotFound
	}
	return &lr.Hosts[0], nil
}

// do is a helper function that makes a request and decodes the response into
// ret.
func do[T any](ret *T, cl *http.Client, r *http.Request) error {
	resp, err := cl.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
		return err
	}
	return nil
}
