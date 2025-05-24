package credentials_provider

import (
	"encoding/json"
	"fmt"

	sp "github.com/sourcehawk/go-flyway/internal/secrets_provider"
)

type AWSSMDatabaseCredentials struct {
	Username *sp.SecretRef `yaml:"username,omitempty"`
	Password *sp.SecretRef `yaml:"password,omitempty"`
	Host     *sp.SecretRef `yaml:"host,omitempty"`
	Port     *sp.SecretRef `yaml:"port,omitempty"`
	Database *sp.SecretRef `yaml:"database,omitempty"`
	awssm    sp.SecretsProvider
}

func (d *AWSSMDatabaseCredentials) Validate() error {
	if d.Username == nil {
		return fmt.Errorf("missing 'username' key in %s credentials", AWSSMSecretsProviderType)
	}
	if d.Password == nil {
		return fmt.Errorf("missing 'password' key in %s credentials", AWSSMSecretsProviderType)
	}
	if d.Host == nil {
		return fmt.Errorf("missing 'host' key in %s credentials", AWSSMSecretsProviderType)
	}
	if d.Port == nil {
		return fmt.Errorf("missing 'port' key in %s credentials", AWSSMSecretsProviderType)
	}
	if d.Database == nil {
		return fmt.Errorf("missing 'database' key in %s credentials", AWSSMSecretsProviderType)
	}
	for _, s := range []*sp.SecretRef{d.Username, d.Password, d.Host, d.Port, d.Database} {
		if err := s.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (d *AWSSMDatabaseCredentials) GetCredentials() (*DatabaseCredentials, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}

	secretsMap := make(map[string]map[string]any)
	credentialsMap := make(map[string]any)

	for _, s := range []sp.SecretRefToStructJsonField{
		{
			StructJsonField: "username",
			SecretRef:       d.Username,
		},
		{
			StructJsonField: "password",
			SecretRef:       d.Password,
		},
		{
			StructJsonField: "host",
			SecretRef:       d.Host,
		},
		{
			StructJsonField: "port",
			SecretRef:       d.Port,
		},
		{
			StructJsonField: "database",
			SecretRef:       d.Database,
		},
	} {
		if err := s.PopulateJSONFieldFromSecret(d.awssm, secretsMap, credentialsMap); err != nil {
			return nil, err
		}
	}

	// convert the map to json
	jsonData, err := json.Marshal(credentialsMap)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal credentials to json: %w", err)
	}

	// convert the json back, this time using a go struct for correct types
	credentials := &DatabaseCredentials{}
	err = json.Unmarshal(jsonData, credentials)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials to json: %w", err)
	}

	return credentials, nil
}
