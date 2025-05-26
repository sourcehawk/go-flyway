package migrator

import (
	"fmt"

	cp "github.com/sourcehawk/go-flyway/internal/credentials_provider"
)

type CredentialProviders struct {
	EnvProviderImpl   *cp.EnvDatabaseCredentials   `yaml:"env,omitempty"`
	TextProviderImpl  *cp.TextDatabaseCredentials      `yaml:"text,omitempty"`
	AwssmProviderImpl *cp.AWSSMDatabaseCredentials `yaml:"aws_sm,omitempty"`
}

type Credentials struct {
	CredentialProviders `yaml:",inline"`
	Provider            string `yaml:"provider"`
	concreteProvider    cp.DatabaseCredentialsProvider
	credentials         *cp.DatabaseCredentials
}

func (c *Credentials) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("missing 'provider' key for database credentials")
	}

	p := cp.CredentialsProviderType(c.Provider)

	switch p {
	case cp.TextProviderType:
		if c.TextProviderImpl == nil {
			return fmt.Errorf("could not find credentials configuration for provider %s", c.Provider)
		}
		c.concreteProvider = c.TextProviderImpl
	case cp.EnvProviderType:
		if c.EnvProviderImpl == nil {
			return fmt.Errorf("could not find credentials configuration for provider %s", c.Provider)
		}
		c.concreteProvider = c.EnvProviderImpl
	case cp.AWSSMSecretsProviderType:
		if c.AwssmProviderImpl == nil {
			return fmt.Errorf("could not find credentials configuration for provider %s", c.Provider)
		}
		c.concreteProvider = c.AwssmProviderImpl
	default:
		return fmt.Errorf("%s is not a valid credentials provider type", c.Provider)
	}

	if err := c.concreteProvider.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *Credentials) fetchCredentials() (*cp.DatabaseCredentials, error) {
	if c.credentials != nil {
		return c.credentials, nil
	}

	creds, err := c.concreteProvider.GetCredentials()

	if err != nil {
		return nil, err
	}

	c.credentials = creds
	return creds, nil
}

// Fetches the credentials from the underlying credentials provider.
// Calls Validate internally
//
// If the credentials have already been fetched, returns existing cached credentials
func (c *Credentials) FetchCredentials() (*cp.DatabaseCredentials, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c.fetchCredentials()
}
