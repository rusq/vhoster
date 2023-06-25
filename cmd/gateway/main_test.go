package main

import (
	"testing"
	"time"

	"github.com/rusq/vhoster"
	"github.com/stretchr/testify/assert"
)

func Test_parseCmdLine(t *testing.T) {
	t.Run("parameters override config values", func(t *testing.T) {
		// setting command line parameters
		addr = ptr("1.2.3.4:8080")
		apiaddr = ptr("5.6.7.8:8083")
		domainName = ptr("example.com")
		config = ptr(writeConfig(t, testConfigJSON))

		cfg, err := parseCmdLine()
		if err != nil {
			t.Fatal(err)
		}
		var (
			wantConfig = &Config{
				GatewayAddress: "1.2.3.4:8080",
				DomainName:     "example.com",
				APIAddress:     "5.6.7.8:8083",
				Timeout:        duration(100 * time.Millisecond),
				Hosts: []vhoster.Host{
					{Name: "vhost.example.com", URI: mustParse("http://localhost:8081")}, // vhost name should have the updated domain name.
				},
			}
		)
		assert.Equal(t, wantConfig, cfg)
	})
}

func ptr[T any](v T) *T {
	return &v
}
