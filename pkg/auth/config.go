package auth

import "fmt"

type Config struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
	Host         string `yaml:"host"`
	SigningKey   string `yaml:"signing_key"`
	Protocol     string `yaml:"protocol"`
}

func (c *Config) Validate(httpEnabled bool) error {
	if c.Host == "" {
		return fmt.Errorf("host de auth é necessário")
	}

	if httpEnabled {
		if c.ClientID == "" {
			return fmt.Errorf("id de cliente de auth é necessário")
		}

		if c.ClientSecret == "" {
			return fmt.Errorf("secret de cliente de auth é necessário")
		}

		if c.RedirectURI == "" {
			return fmt.Errorf("uri de redirect de auth é necessário")
		}

		if c.SigningKey == "" {
			return fmt.Errorf("signing key de auth é necessária")
		}

		if c.Protocol != "http" && c.Protocol != "https" {
			return fmt.Errorf("protocolo de auth é necessário. valores aceitos são http e https")
		}
	}

	return nil
}
