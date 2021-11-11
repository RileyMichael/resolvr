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

	// server address + port
	BindAddress string `envconfig:"default=:53"`

	// metrics address + port
	MetricsAddress string `envconfig:"default=:9091"`

	// dev / prod
	Env string `envconfig:"default=dev"`

	// First: Host Second Address
	StaticTypeARecords []StaticConfig `envconfig:"default={resolvr.io.;127.0.0.1}"`

	// First: Host Second: Address
	StaticTypeAAAARecords []StaticConfig `envconfig:"default={resolvr.io.;::1}"`

	// First: Alias Second: Canonical
	StaticTypeCNAMERecords []StaticConfig `envconfig:"default={www.resolvr.io.;resolvr.io.}"`

	// First: Host, Second: Address
	// NS records will be created for root Hostname to every NS Host, and A records for NS Host -> NS Address
	Nameservers []StaticConfig `envconfig:"default={ns1.resolvr.io.;127.0.0.1};{ns2.resolvr.io.;127.0.0.1}"`
}

type StaticConfig struct {
	First  string
	Second string
}
