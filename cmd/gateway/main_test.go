package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_parseCmdLine(t *testing.T) {
	t.Run("parameters override config values", func(t *testing.T) {
		// setting command line parameters
		addr = ptr("1.2.3.4:8080")
		apiaddr = ptr("5.6.7.8:8083")
		pubAddr = ptr("example.com")
		config = ptr(writeConfig(t, testConfigJSON))

		cfg, hosts, err := parseCmdLine()
		if err != nil {
			t.Fatal(err)
		}
		var (
			wantConfig = &Config{
				GatewayAddress: "1.2.3.4:8080",
				DomainName:     "example.com",
				APIAddress:     "5.6.7.8:8083",
				Timeout:        duration(100 * time.Millisecond),
				StrHosts:       testCfg.StrHosts,
			}
			wantHosts, _ = testCfg.Hosts()
		)
		assert.Equal(t, wantConfig, cfg)
		assert.Equal(t, wantHosts, hosts)
	})
}

func ptr[T any](v T) *T {
	return &v
}
