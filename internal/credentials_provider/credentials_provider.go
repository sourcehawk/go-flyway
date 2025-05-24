package credentials_provider

import "fmt"

type CredentialsProviderType string

const (
	TextProviderType         CredentialsProviderType = "text"
	EnvProviderType          CredentialsProviderType = "env"
	AWSSMSecretsProviderType CredentialsProviderType = "aws_sm"
)

type DatabaseCredentialsProvider interface {
	// Validates the struct for any errors / missing fields etc
	Validate() error
	// Returns database credentials according to the given configuration
	GetCredentials() (*DatabaseCredentials, error)
}

type DatabaseCredentials struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Host     string `json:"host,omitempty" yaml:"host,omitempty"`
	Port     int    `json:"port,omitempty" yaml:"port,omitempty"`
	Database string `json:"database,omitempty" yaml:"database,omitempty"`
}

func (d *DatabaseCredentials) Validate() error {
	if d.Username == "" {
		return fmt.Errorf("missing 'username' key in database credentials")
	}
	if d.Password == "" {
		return fmt.Errorf("missing 'password' key in database credentials")
	}
	if d.Host == "" {
		return fmt.Errorf("missing 'host' key in database credentials")
	}
	if d.Port == 0 {
		return fmt.Errorf("missing 'port' key in database credentials")
	}
	if d.Database == "" {
		return fmt.Errorf("missing 'database' key in database credentials")
	}
	return nil
}

func (d *DatabaseCredentials) GetCredentials() (*DatabaseCredentials, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return d, nil
}
