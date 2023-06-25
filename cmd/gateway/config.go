package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rusq/vhoster"
)

type duration time.Duration

type Config struct {
	GatewayAddress string         `json:"gateway_address,omitempty"`
	DomainName     string         `json:"domain_name,omitempty"`
	APIAddress     string         `json:"api_address,omitempty"`
	Timeout        duration       `json:"timeout,omitempty"`
	Hosts          []vhoster.Host `json:"hosts,omitempty"`
}

func (c *Config) validate() error {
	if c.GatewayAddress == "" {
		return errors.New("gateway address is empty")
	}
	if c.DomainName == "" {
		return errors.New("domain name is empty")
	}
	if c.APIAddress == "" {
		return errors.New("api address is empty")
	}
	for i, h := range c.Hosts {
		if err := h.Validate(); err != nil {
			return fmt.Errorf("error validating configuration host %d: %w", i, err)
		}
	}
	return nil
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

func loadConfig(path string, cfg *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()

	if err := dec.Decode(cfg); err != nil {
		return err
	}

	return cfg.validate()
}
