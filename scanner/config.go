package main

import (
	"fmt"
	"os"

	"github.com/elmasy-com/elnet/ctlog"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogName       string     `yaml:"LogName"`
	MongoURI      string     `yaml:"MongoURI"`
	InsertWorkers int        `yaml:"InsertWorkers"`
	Log           *ctlog.Log `yaml:"-"`
}

var Conf *Config

// ParseConfig parses the config file in path and set the global variable Conf.
func ParseConfig(path string) error {

	Conf = &Config{}

	out, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %s", path, err)
	}

	err = yaml.Unmarshal(out, Conf)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %s", err)
	}

	switch {
	case Conf.LogName == "":
		return fmt.Errorf("LogName is missing")
	case Conf.MongoURI == "":
		return fmt.Errorf("MongoURI is missing")
	}

	Conf.Log = ctlog.LogByName(Conf.LogName)
	if Conf.Log == nil {
		return fmt.Errorf("unknown LogName: %s", Conf.LogName)
	}

	if Conf.InsertWorkers < 0 {
		return fmt.Errorf("NumWorkers is negative")
	}
	if Conf.InsertWorkers == 0 {
		Conf.InsertWorkers = 2
	}

	return nil
}
