package config

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/elmasy-com/elnet/dns"
	"gopkg.in/yaml.v3"
)

type conf struct {
	MongoURI       string   `yaml:"MongoURI"`
	Address        string   `yaml:"Address"`
	TrustedProxies []string `yaml:"TrustedProxies"`
	SSLCert        string   `yaml:"SSLCert"`
	SSLKey         string   `yaml:"SSLKey"`
	LogErrorOnly   bool     `yaml:"LogErrorOnly"`
	DNSServers     []string `yaml:"DNSServers"`
	DomainWorker   int      `yaml:"DomainWorker"`
	DomainBuffer   int      `yaml:"DomainBuffer"`
}

var (
	MongoURI       string   // MongoDB connection string
	Address        string   // Address to listen on
	TrustedProxies []string // A list of trusted proxies
	SSLCert        string
	SSLKey         string
	LogErrorOnly   bool
	DNSServers     []string
	DomainWorker   int
	DomainBuffer   int
)

// Parse parses the config file in path and gill the global variables.
func Parse(path string) error {

	c := conf{}

	out, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %s", path, err)
	}

	err = yaml.Unmarshal(out, &c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %s", err)
	}

	if c.MongoURI == "" {
		return fmt.Errorf("MongoURI is empty")
	}
	MongoURI = c.MongoURI

	if c.Address == "" {
		c.Address = ":8080"
	}
	Address = c.Address

	TrustedProxies = c.TrustedProxies

	SSLCert = c.SSLCert
	SSLKey = c.SSLKey

	LogErrorOnly = c.LogErrorOnly

	DNSServers = c.DNSServers

	servers, err := dns.NewServersStr(dns.DefaultMaxRetries, time.Duration(dns.DefaultQueryTimeoutSec)*time.Second, c.DNSServers...)
	if err != nil {
		return fmt.Errorf("failed to update DNS servers: %w", err)
	}

	dns.DefaultServers = servers

	if c.DomainWorker == 0 {
		c.DomainWorker = runtime.NumCPU()
	}

	DomainWorker = c.DomainWorker

	if c.DomainBuffer == 0 {
		c.DomainBuffer = 1000
	}

	DomainBuffer = c.DomainBuffer

	return nil
}
