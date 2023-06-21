package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rusq/vhoster/apiserver"
)

var (
	vhostsURL = &url.URL{Path: "/vhost/"}
)

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
	req, err := http.NewRequest(http.MethodPost, c.base.ResolveReference(vhostsURL).String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.cl.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var addResp apiserver.AddResponse
	if err := json.NewDecoder(resp.Body).Decode(&addResp); err != nil {
		return "", err
	}
	return addResp.Hostname, nil
}

func (c *Client) Remove(hostname string) error {
	url := &url.URL{Path: "/vhost/" + hostname}
	req, err := http.NewRequest(http.MethodDelete, c.base.ResolveReference(url).String(), nil)
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
	req, err := http.NewRequest(http.MethodGet, c.base.ResolveReference(vhostsURL).String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var listResp apiserver.ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}
	return listResp.Hosts, nil
}
