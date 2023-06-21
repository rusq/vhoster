package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"github.com/rusq/vhoster/apiserver"
)

type Client struct {
	base *url.URL
	cl   *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{base: u, cl: httpClient}, nil
}


func (c *Client) Add(hostPrefix, target string) (string, error) {
	reqBody, err := json.Marshal(apiserver.AddRequest{
		HostPrefix: hostPrefix,
		Target: target,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, c.base.String()+"/vhost/", bytes.NewBuffer(reqBody))
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
	req, err := http.NewRequest(http.MethodDelete, c.base.String()+"/vhost/"+hostname, nil)
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
	req, err := http.NewRequest(http.MethodGet, c.base.String()+"/vhost/", nil)
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
