package config

import (
	"os"
	"time"

	"github.com/go-yaml/yaml"
)

// TO DO
type Config struct {
	Interval time.Duration `yaml:"interval"`
	Server   struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Metrics struct {
		EnableLoadAverage     bool `yaml:"enableLoadAverage"`
		EnableCPU             bool `yaml:"enableCPU"`
		EnableDiskUsage       bool `yaml:"enableDiskUsage"`
		EnableFileSystemUsage bool `yaml:"enableFileSystemUsage"`
		EnableNetworkProtocol bool `yaml:"enableNetworkProtocol"`
	} `yaml:"metrics"`
}

func LoadConf(path string) (*Config, error) {
	var newConf Config
	byteConf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(byteConf, &newConf); err != nil {
		return nil, err
	}
	return &newConf, nil
}
