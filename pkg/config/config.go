package config

import (
	"fmt"

	"github.com/keepchasingheaven/encore-workspaces-proxy/pkg/auth"
)

type HTTP struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type SSH struct {
	Enabled         bool   `yaml:"enabled"`
	Port            int    `yaml:"port"`
	HostKey         string `yaml:"host_key"`
	BackendPort     int    `yaml:"backend_port"`
	BackendUsername string `yaml:"backend_username"`
}

type Config struct {
	Auth            auth.Config `yaml:"auth"`
	MetricsPath     string      `yaml:"metrics_path"`
	LogLevel        string      `yaml:"log_level"`
	HTTP            HTTP        `yaml:"http"`
	SSH             SSH         `yaml:"ssh"`
	WatchNamespaces []string    `yaml:"watch_namespaces"`
}

func (h *HTTP) Validate() error {
	if h.Enabled {
		if h.Port == 0 {
			return fmt.Errorf("porta de http é necessária")
		}
	}
	
	return nil
}

func (s *SSH) Validate() error {
	if s.Enabled {
		if s.Port == 0 {
			return fmt.Errorf("porta de ssh é necessária")
		}
		
		if s.HostKey == "" {
			return fmt.Errorf("chave de host de ssh é necessária")
		}
		
		if s.BackendPort == 0 {
			return fmt.Errorf("porta de backend de ssh é necessária")
		}
		
		if s.BackendUsername == "" {
			return fmt.Errorf("nome de usuário de backend de ssh é necessário")
		}
	}
	
	return nil
}

func (c *Config) Validate() error {
	err := c.Auth.Validate(c.HTTP.Enabled)
	if err != nil {
		return err
	}

	if c.MetricsPath == "" {
		return fmt.Errorf("path de métricas é necessário")
	}

	if c.LogLevel != "info" && c.LogLevel != "debug" && c.LogLevel != "warn" && c.LogLevel != "error" {
		return fmt.Errorf(
			"nível de log é necessário. valores aceitos: info, debug, warn e error",
		)
	}

	err = c.HTTP.Validate()
	if err != nil {
		return err
	}

	err = c.SSH.Validate()
	if err != nil {
		return err
	}

	return nil
}
