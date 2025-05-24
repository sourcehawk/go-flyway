package credentials_provider

import (
	"fmt"
	"os"
	"strconv"
)

type EnvDatabaseCredentials struct {
	UsernameKey string `yaml:"usernameKey"`
	PasswordKey string `yaml:"passwordKey"`
	HostKey     string `yaml:"hostKey"`
	PortKey     string `yaml:"portKey"`
	DatabaseKey string `yaml:"databaseKey"`

	username string
	password string
	host     string
	port     int
	database string
}

func (e *EnvDatabaseCredentials) nonEmptyEnvOrError(key string) (string, error) {
	keyValue, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("environment variable %s specified in %s credentials not set", key, EnvProviderType)
	}
	if keyValue == "" {
		return "", fmt.Errorf("environment variable %s specified in %s credentials has empty value", key, EnvProviderType)
	}
	return keyValue, nil
}

func (e *EnvDatabaseCredentials) loadFromEnv() error {
	var err error

	if e.username, err = e.nonEmptyEnvOrError(e.UsernameKey); err != nil {
		return err
	}
	if e.password, err = e.nonEmptyEnvOrError(e.PasswordKey); err != nil {
		return err
	}
	if e.host, err = e.nonEmptyEnvOrError(e.HostKey); err != nil {
		return err
	}
	if e.database, err = e.nonEmptyEnvOrError(e.DatabaseKey); err != nil {
		return err
	}

	portStr, err := e.nonEmptyEnvOrError(e.PortKey)

	if err != nil {
		return err
	}

	if e.port, err = strconv.Atoi(portStr); err != nil {
		return err
	}

	return nil
}

func (e *EnvDatabaseCredentials) Validate() error {
	if e.UsernameKey == "" {
		return fmt.Errorf("missing 'usernameKey' in %s credentials", EnvProviderType)
	}
	if e.PasswordKey == "" {
		return fmt.Errorf("missing 'passwordKey' in %s credentials", EnvProviderType)
	}
	if e.HostKey == "" {
		return fmt.Errorf("missing 'hostKey' in %s credentials", EnvProviderType)
	}
	if e.PortKey == "" {
		return fmt.Errorf("missing 'portKey' in %s credentials", EnvProviderType)
	}
	if e.DatabaseKey == "" {
		return fmt.Errorf("missing 'databaseKey in %s credentials", EnvProviderType)
	}
	if err := e.loadFromEnv(); err != nil {
		return err
	}
	return nil
}

func (e *EnvDatabaseCredentials) GetCredentials() (*DatabaseCredentials, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return &DatabaseCredentials{
		Username: e.username,
		Password: e.password,
		Host:     e.host,
		Port:     e.port,
		Database: e.database,
	}, nil
}
