package main

import (
	"flag"
	"log"
	"time"

	"github.com/rusq/osenv/v2"
	"github.com/rusq/vhoster"
	"github.com/rusq/vhoster/apiserver"
)

var (
	addr       = flag.String("addr", osenv.Value("GATEWAY_ADDRESS", ""), "gateway address (host:port)")
	domainName = flag.String("domain", osenv.Value("DOMAIN", ""), "server public domain `name`, it is used as a suffix for all vhosts, e.g. vhost1.public-hostname.com.  It must include custom port, if it uses one.")
	apiaddr    = flag.String("api", osenv.Value("API_ADDRESS", ""), "address of this api server that controls the gateway")
	config     = flag.String("c", osenv.Value("CONFIG", ""), "path to the optional config file in JSON format.")
)

func main() {
	flag.Parse()

	cfg, err := parseCmdLine()
	if err != nil {
		log.Fatal(err)
	}

	s, err := vhoster.Listen(cfg.GatewayAddress, vhoster.WithHosts(cfg.Hosts), vhoster.WithTimeout(time.Duration(cfg.Timeout)))
	if err != nil {
		log.Fatal(err)
	}
	go s.Wait()
	log.Printf("gateway started on %s ; API adddress: %s", cfg.GatewayAddress, cfg.APIAddress)
	log.Fatal(apiserver.Run(s, cfg.APIAddress, cfg.DomainName))
}

func parseCmdLine() (*Config, error) {
	var cfg = Config{}
	if *config != "" {
		if err := loadConfig(*config, &cfg); err != nil {
			return nil, err
		}
	}
	// override config with command line flags.
	cfg.GatewayAddress = coalesce(*addr, cfg.GatewayAddress)
	cfg.APIAddress = coalesce(*apiaddr, cfg.APIAddress)
	cfg.DomainName = coalesce(*domainName, cfg.DomainName)
	if cfg.Timeout == 0 {
		cfg.Timeout = duration(5 * time.Second)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	for i, h := range cfg.Hosts {
		cfg.Hosts[i].Name = h.Name + "." + cfg.DomainName
	}

	return &cfg, nil
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
