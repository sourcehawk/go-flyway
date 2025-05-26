package credentials_provider

import "fmt"

type TextDatabaseCredentials struct {
	DatabaseCredentials `yaml:",inline"`
}

func (d *TextDatabaseCredentials) Validate() error {
	if d.Username == "" {
		return fmt.Errorf("missing 'username' key in %s credentials", TextProviderType)
	}
	if d.Password == "" {
		return fmt.Errorf("missing 'password' key in %s credentials", TextProviderType)
	}
	if d.Host == "" {
		return fmt.Errorf("missing 'host' key in %s credentials", TextProviderType)
	}
	if d.Port == 0 {
		return fmt.Errorf("missing 'port' key in %s credentials", TextProviderType)
	}
	if d.Database == "" {
		return fmt.Errorf("missing 'database' key in %s credentials", TextProviderType)
	}
	return nil
}

func (d *TextDatabaseCredentials) GetCredentials() (*DatabaseCredentials, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return &d.DatabaseCredentials, nil
}
