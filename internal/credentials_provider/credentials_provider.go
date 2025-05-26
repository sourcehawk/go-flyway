package credentials_provider

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

