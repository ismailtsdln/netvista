package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Ports       string `yaml:"ports"`
	Concurrency int    `yaml:"concurrency"`
	Output      string `yaml:"output"`
	Timeout     string `yaml:"timeout"`
	Proxy       string `yaml:"proxy"`
	Headers     string `yaml:"headers"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{
		Ports:       "80,443",
		Concurrency: 20,
		Output:      "./out",
		Timeout:     "5s",
	}

	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(data, config)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
