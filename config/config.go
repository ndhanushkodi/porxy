package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Listeners []Listener `yaml:"listeners"`
	Backends  []Backend  `yaml:"backends"`
}

type Listener struct {
	Name    string `yaml:"name"`
	Backend string `yaml:"backend"`
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}

type Backend struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func LoadConfig(rawconfig []byte) Config {

	config := Config{}
	err := yaml.Unmarshal(rawconfig, &config)
	if err != nil {
		fmt.Println("Failed to unmarshal config")
		os.Exit(1)
	}
	return config
}

func (c Config) GetBackend(backend string) Backend {
	for _, b := range c.Backends {
		if b.Name == backend {
			return b
		}
	}
	// TODO handle this case with err
	return Backend{}
}
