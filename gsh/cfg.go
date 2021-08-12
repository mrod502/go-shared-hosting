package gsh

type Config struct {
	DomainConfigPath string `yaml:"domain_config_path,omitempty"`
	ServePort        uint16 `yaml:"serve_port,omitempty"`
	LoggerAddress    string `yaml:"logger_address,omitempty"`
	LogPrefix        string `yaml:"log_prefix,omitempty"`
	TLS              bool   `yaml:"tls,omitempty"`
	CertFile         string `yaml:"cert_file,omitempty"`
	KeyFile          string `yaml:"key_file,omitempty"`
}
