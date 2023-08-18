package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Resolvers     []string `yaml:"Resolvers"`
	MongoURI      string   `yaml:"MongoURI"`
	NumWorkers    int      `yaml:"NumWorkers"`
	BuffSize      int      `yaml:"BuffSize"`
	ListenAddress string   `yaml:"ListenAddress"`
}

// parseConfig parses the config file in path, set the default if needed and return the Config struct.
func parseConfig(path string) (Config, error) {

	c := Config{}

	out, err := os.ReadFile(path)
	if err != nil {
		return c, fmt.Errorf("failed to read %s: %w", path, err)
	}

	err = yaml.Unmarshal(out, &c)
	if err != nil {
		return c, fmt.Errorf("failed to unmarshal: %w", err)
	}

	if len(c.Resolvers) == 0 {
		c.Resolvers = []string{"1.1.1.1:53", "1.0.0.1:53"}
	}
	for i := range c.Resolvers {
		if !strings.Contains(c.Resolvers[i], ":") {
			return c, fmt.Errorf("missing port from %s", c.Resolvers[i])
		}
	}

	if c.MongoURI == "" {
		return c, fmt.Errorf("MongoURI is missing")
	}

	if c.NumWorkers <= 0 {
		c.NumWorkers = 4
	}

	if c.BuffSize <= 0 {
		c.BuffSize = 1000
	}

	if c.ListenAddress == "" {
		c.ListenAddress = ":1053"
	}

	return c, nil
}
