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
	addr    = flag.String("addr", osenv.Value("GATEWAY_ADDRESS", "localhost:8080"), "gateway address (host:port)")
	pubAddr = flag.String("domain", osenv.Value("DOMAIN", "localhost:8080"), "server public domain `name`, it is used as a suffix for all vhosts, e.g. vhost1.public-hostname.com.  It must include custom port, if it uses one.")
	apiaddr = flag.String("api", osenv.Value("API_ADDRESS", "localhost:8083"), "address of this api server that controls the gateway")
	config  = flag.String("c", osenv.Value("CONFIG", ""), "path to the optional config file in JSON format.")
)

func main() {
	flag.Parse()

	cfg, hosts, err := parseCmdLine()
	if err != nil {
		log.Fatal(err)
	}

	s, err := vhoster.Listen(cfg.GatewayAddress, vhoster.WithHosts(hosts), vhoster.WithTimeout(time.Duration(cfg.Timeout)))
	if err != nil {
		log.Fatal(err)
	}
	go s.Wait()

	log.Fatal(apiserver.Run(s, cfg.APIAddress, cfg.DomainName))
}

func parseCmdLine() (*Config, []vhoster.Host, error) {
	var cfg = Config{}
	if *config != "" {
		if err := loadConfig(*config, &cfg); err != nil {
			log.Fatal(err)
		}
	}
	// override config with command line flags.
	cfg.GatewayAddress = coalesce(*addr, cfg.GatewayAddress)
	cfg.APIAddress = coalesce(*apiaddr, cfg.APIAddress)
	cfg.DomainName = coalesce(*pubAddr, cfg.DomainName)
	if cfg.Timeout == 0 {
		cfg.Timeout = duration(5 * time.Second)
	}

	hosts, err := cfg.Hosts()
	if err != nil {
		log.Fatal(err)
	}
	return &cfg, hosts, nil
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
