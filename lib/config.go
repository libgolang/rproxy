package lib

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// Config is used to configure instances of ReverseProxy.  Used
// as argument of NewReverseProxy.  It represents the configuration
// from the TOML file
type Config struct {
	Ip                  *string       `toml:"ip"`                  // Default 0.0.0.0
	Port                *int          `toml:"port"`                // Default 8080
	MaxIdleConns        *int          `toml:"maxIdleConns"`        // Default 10000
	MaxIdleConnsPerHost *int          `toml:"maxIdleConnsPerHost"` // Default 10000
	Proxies             []ConfigProxy `toml:"proxy"`
}

// ConfigProxy Represnets instances of `[[proxy]]` sections in the
// TOML configuration file. The `DomainName` field is the incomfing host
// part of the URL and the `Remote` field is a list of URLs to forward
// the requests to.
type ConfigProxy struct {
	DomainName string   `toml:"domainName"` // example.com
	Remote     []string `toml:"remote"`     // ["http://service1.example.com:8080","http://service2.example.com"]
}

// ReadConfig loads configuration from a TOML file into an instance
// of `Config` struct
func ReadConfig(file string) (*Config, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file %s: %s", file, err)
	}
	cfg := &Config{}
	if err := toml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("Error reading config file %s: %s", file, err)
	}

	if cfg.Ip == nil {
		*cfg.Ip = "0.0.0.0"
	}
	if cfg.MaxIdleConns == nil {
		*cfg.MaxIdleConns = 10000
	}
	if cfg.MaxIdleConnsPerHost == nil {
		*cfg.MaxIdleConnsPerHost = 10000
	}
	if cfg.Port == nil {
		*cfg.Port = 8080
	}
	return cfg, nil
}
