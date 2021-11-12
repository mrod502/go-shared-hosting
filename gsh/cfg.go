package gsh

import "github.com/mrod502/logger"

type Config struct {
	logger.ClientConfig
	RoutesFile    string `yaml:"routes_file"`
	ServePort     uint16 `yaml:"serve_port,omitempty"`
	LoggerAddress string `yaml:"logger_address,omitempty"`
	TLS           bool   `yaml:"tls,omitempty"`
	CertFile      string `yaml:"cert_file,omitempty"`
	KeyFile       string `yaml:"key_file,omitempty"`
}
