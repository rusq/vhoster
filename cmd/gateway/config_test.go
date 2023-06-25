package main

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/rusq/vhoster"
	"github.com/stretchr/testify/assert"
)

const testConfigJSON = `
{
	"gateway_address": "0.0.0.0:8080",
	"api_address": "0.0.0.0:8083",
	"domain_name": "localhost:8080",
	"timeout": "100ms",
	"hosts": [
		{
			"name": "vhost",
			"uri": "http://localhost:8081"
		}
	]
}
`

func mustParse(s string) *vhoster.URI {
	uri, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return vhoster.ToURI(uri)
}

var testCfg = &Config{
	GatewayAddress: "0.0.0.0:8080",
	DomainName:     "localhost:8080",
	APIAddress:     "0.0.0.0:8083",
	Timeout:        duration(100 * time.Millisecond),
	Hosts: []vhoster.Host{
		{Name: "vhost", URI: mustParse("http://localhost:8081")},
	},
}

func writeConfig(t *testing.T, configjs string) string {
	t.Helper()
	d := t.TempDir()
	f, err := os.CreateTemp(d, "config.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(configjs); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func Test_loadConfig(t *testing.T) {
	testcfg := writeConfig(t, testConfigJSON)
	type args struct {
		path string
		cfg  *Config
	}
	tests := []struct {
		name       string
		args       args
		wantConfig *Config
		wantErr    bool
	}{
		{
			"valid config",
			args{
				testcfg,
				&Config{},
			},
			testCfg,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := loadConfig(tt.args.path, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantConfig != nil {
				assert.Equal(t, tt.wantConfig, tt.args.cfg)
			}
		})
	}
}
