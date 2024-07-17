package config

import (
	"os"
	"time"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Interval time.Duration `yaml:"interval"`
	Server   struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
}

func (c Config) LoadConf(path string) (*Config, error) {
	var newConf Config

	byteConf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(byteConf, newConf); err != nil {
		return nil, err
	}
	return &newConf, nil
}
