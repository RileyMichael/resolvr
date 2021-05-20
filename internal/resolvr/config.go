package resolvr

import (
	"github.com/vrischmann/envconfig"
)

func LoadConfig() (*Config, error) {
	config := Config{}
	if err := envconfig.InitWithPrefix(&config, "resolvr"); err != nil {
		return nil, err
	}
	return &config, nil
}

type Config struct {
	// the base hostname with a trailing dot
	Hostname string `envconfig:"default=resolvr.io."`

	// address for base hostname A record
	Address string `envconfig:"default=127.0.0.1"`

	// server address + port
	BindAddress string `envconfig:"default=:53"`

	// dev / prod
	Env string `envconfig:"default=dev"`

	// slice of nameservers
	Nameserver []NameserverConfig `envconfig:"default={ns1.resolvr.io.;127.0.0.1};{ns2.resolvr.io.;127.0.0.1}"`
}

type NameserverConfig struct {
	// full hostname for the nameserver with a trailing dot
	Hostname string

	// ip address for the A record
	Address string
}
