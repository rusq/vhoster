package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/rusq/vhoster"
)

type duration time.Duration

type Config struct {
	GatewayAddress string   `json:"gateway_address,omitempty"`
	DomainName     string   `json:"domain_name,omitempty"`
	APIAddress     string   `json:"api_address,omitempty"`
	Timeout        duration `json:"timeout,omitempty"`
	StrHosts       []Host   `json:"hosts,omitempty"`
}

func (d *duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	td, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = duration(td)
	return nil
}

func (c Config) Hosts() ([]vhoster.Host, error) {
	var hs []vhoster.Host
	for i, h := range c.StrHosts {
		vh, err := h.toVhost(c.DomainName)
		if err != nil {
			return nil, fmt.Errorf("failed on host #%d: %w", i+1, err)
		}
		hs = append(hs, vh)
	}
	return hs, nil
}

type Host struct {
	Subdomain string `json:"subdomain,omitempty"`
	URL       string `json:"url,omitempty"`
}

func (h Host) validate() error {
	if h.URL == "" {
		return errors.New("host with empty url")
	}
	if h.Subdomain == "" {
		return fmt.Errorf("host with empty subdomain for url: %s", h.URL)
	}
	return nil
}

func (h Host) toVhost(domain string) (vhoster.Host, error) {
	if err := h.validate(); err != nil {
		return vhoster.Host{}, err
	}
	u, err := url.Parse(h.URL)
	if err != nil {
		return vhoster.Host{}, err
	}
	return vhoster.Host{
		Name: h.Subdomain + "." + domain,
		URI:  u,
	}, nil
}

func loadConfig(path string, cfg *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()

	return dec.Decode(cfg)
}
