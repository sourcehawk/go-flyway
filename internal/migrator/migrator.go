package migrator

import (
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/yaml.v3"
)

type Migrator struct {
	// Flyway arguments applied globally
	FlywayArgs []string `yaml:"flywayArgs,omitempty"`
	// Credentials applied globally to schemas unless they explicitly specify their own
	Credentials *Credentials `yaml:"credentials,omitempty"`
	// List of schemas to migrate
	Schemas     []*Schema `yaml:"schemas"`
	cmdExecFunc CommandFuncType
}

// Validate that the migrator configuration is valid
// Note that this only validates the structure of the configuration,
// it does not mean that the migration command will succceed
func (m *Migrator) Validate() error {
	if m.Credentials != nil {
		if err := m.Credentials.Validate(); err != nil {
			return err
		}
		// prefetch credentials to fail fast instead of during migration process
		if _, err := m.Credentials.FetchCredentials(); err != nil {
			return err
		}
	}

	for _, s := range m.Schemas {
		if s.Credentials == nil {
			if m.Credentials == nil {
				return fmt.Errorf("missing 'credentials' field in migrator config for schema %s", s.Name)
			}
			s.Credentials = m.Credentials
		}

		if len(m.FlywayArgs) != 0 {
			if err := s.SetDefaultFlywayArgs(m.FlywayArgs); err != nil {
				return err
			}
		}

		if err := s.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Run the migrator according to it's configuration
func (m *Migrator) Migrate() error {
	if err := m.Validate(); err != nil {
		return err
	}

	for _, s := range m.Schemas {
		if err := s.Migrate(m.cmdExecFunc); err != nil {
			return err
		}
	}

	return nil
}

func newMigrator(configFile string, cmdExecFn CommandFuncType) (*Migrator, error) {
	data, err := os.ReadFile(configFile)

	if err != nil {
		return nil, fmt.Errorf("unable to read config file %s: %w", configFile, err)
	}

	var execFn CommandFuncType = exec.Command

	if cmdExecFn != nil {
		execFn = cmdExecFn
	}

	migrator := &Migrator{
		cmdExecFunc: execFn,
	}

	if err := yaml.Unmarshal(data, migrator); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config file %s to migrator: %w", configFile, err)
	}

	if err := migrator.Validate(); err != nil {
		return nil, err
	}

	return migrator, nil
}

// Create a new migrator from a config file
func NewMigrator(configFile string) (*Migrator, error) {
	return newMigrator(configFile, nil)
}
