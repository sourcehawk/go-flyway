package migrator

import (
	"fmt"
	"os"
)

type Placeholder struct {
	// Name of the placeholder value that shall be replaced
	Name string `yaml:"name"`
	// The value that shall be put in place of the placeholder
	Value string `yaml:"value,omitempty"`
	// Optionally, the user can load the value from a given file path
	ValueFromFile string `yaml:"valueFromFile,omitempty"`
}

func (p *Placeholder) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("'name' cannot be empty in placeholder value")
	}

	if p.Value == "" && p.ValueFromFile == "" {
		return fmt.Errorf("must specify either 'value' or 'valueFromFile', both empty for %s", p.Name)
	}

	return nil
}

func (p *Placeholder) loadValueFromFile() error {
	if p.ValueFromFile == "" {
		panic("ValueFromFile not set, cannot read file")
	}

	data, err := os.ReadFile(p.ValueFromFile)

	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	p.Value = string(data)

	if p.Value == "" {
		return fmt.Errorf("value empty after loading from file %s", p.ValueFromFile)
	}

	return nil
}

func (p *Placeholder) ToFlywayArg() (string, error) {
	err := p.Validate()
	if err != nil {
		return "", err
	}

	if p.ValueFromFile != "" {
		if err := p.loadValueFromFile(); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("-placeholders.%s=%s", p.Name, p.Value), nil
}
