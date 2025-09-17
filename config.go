package main

import (
	"log"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

var cfg *Config

type Config struct {
	Fork         bool     `hcl:"fork,optional"`
	ProxyEnabled bool     `hcl:"proxy_enabled,optional"`
	LogLevel     string   `hcl:"log_level,optional"`
	Listen       string   `hcl:"listen,optional"`
	Proxy        string   `hcl:"proxy,optional"`
	Rules        []string `hcl:"rules"`
}

func ParseConfig(p string) (*Config, error) {
	var c Config
	err := hclsimple.DecodeFile(p, nil, &c)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)

		return nil, err
	}

	if c.LogLevel == "" {
		c.LogLevel = "info"
	}

	if c.Listen == "" {
		c.Listen = ":3000"
	}

	return &c, nil
}
