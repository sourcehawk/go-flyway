package migrator

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CommandFuncType func(name string, arg ...string) *exec.Cmd

type Schema struct {
	// Name of the schema
	Name string `yaml:"name"`
	// Path where the migration files live
	MigrationsPath string `yaml:"migrationsPath"`
	// Arguments to pass to flyway
	FlywayArgs []string `yaml:"flywayArgs,omitempty"`
	// Placeholders for the migrations in the schema
	Placeholders []*Placeholder `yaml:"placeholders,omitempty"`
	// Database credentials
	Credentials *Credentials `yaml:"credentials,omitempty"`
}

func (s *Schema) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("missing 'name' in schema")
	}

	if s.MigrationsPath == "" {
		return fmt.Errorf("missing 'migrationsPath' in schema")
	}

	if s.Credentials == nil {
		return fmt.Errorf("missing credentials for schema %s", s.Name)
	}

	// prefetch so that we get an error at config load time
	// in case of problematic config
	if _, err := s.Credentials.FetchCredentials(); err != nil {
		return err
	}

	for _, arg := range s.FlywayArgs {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			return fmt.Errorf(
				"flyway argument %s cannot be interpreted. "+
					"Ensure format is key=value with no extra '='",
				arg,
			)
		}
		if len(kv[0]) < 2 || len(kv[1]) < 1 {
			return fmt.Errorf(
				"flyway argument %s cannot be interpreted. "+
					"Ensure format is -key=value",
				arg,
			)
		}
		if kv[0][:1] != "-" {
			return fmt.Errorf(
				"flyway argument %s cannot be interpreted. "+
					"Must start with a dash (-)",
				arg,
			)
		}
	}

	for _, p := range s.Placeholders {
		if err := p.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Add arguments to the current flyway args, retaining the current ones if there is a key clash
//
// Ignores any invalid arguments
func (s *Schema) SetDefaultFlywayArgs(args []string) error {
	// Note that it is important that the flyway args are the tail of the array
	// Otherwise the default args would overwrite the concrete ones for the schema
	allArgs := append(args, s.FlywayArgs...)

	set := make(map[string]string, len(s.FlywayArgs))

	for _, arg := range allArgs {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			return fmt.Errorf(
				"flyway argument %s cannot be interpreted. "+
					"Ensure format is key=value with no extra '='",
				arg,
			)
		}
		for _, arg := range s.FlywayArgs {
			kv := strings.Split(arg, "=")
			if len(kv) != 2 {
				return fmt.Errorf(
					"flyway argument %s cannot be interpreted. "+
						"Ensure format is -key=value with no extra '='",
					arg,
				)
			}
		}
		set[kv[0]] = kv[1]
	}

	newArgs := make([]string, len(set))
	idx := 0

	for key, value := range set {
		newArgs[idx] = fmt.Sprintf("%s=%s", key, value)
		idx++
	}

	s.FlywayArgs = newArgs
	return nil
}

func (s *Schema) ensureFlyway(commandExecutor CommandFuncType) error {
	cmd := commandExecutor("flyway")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("flyway not installed, please install before trying again: %w", err)
	}
	return nil
}

func (s *Schema) Migrate(commandExecutor CommandFuncType) error {
	if err := s.Validate(); err != nil {
		return err
	}

	if err := s.ensureFlyway(commandExecutor); err != nil {
		return err
	}

	creds, err := s.Credentials.FetchCredentials()
	if err != nil {
		return err
	}

	allArgs := []string{}
	allArgs = append(allArgs, s.FlywayArgs...)

	for _, p := range s.Placeholders {
		pArg, err := p.ToFlywayArg()
		if err != nil {
			return err
		}
		allArgs = append(allArgs, pArg)
	}

	defaultArgs := []string{
		fmt.Sprintf("-user=%s", creds.Username),
		fmt.Sprintf("-password=%s", creds.Password),
		fmt.Sprintf("-url=jdbc:postgresql://%s:%d/%s", creds.Host, creds.Port, creds.Database),
		fmt.Sprintf("-schemas=%s", s.Name),
		fmt.Sprintf("-locations=filesystem:%s", s.MigrationsPath),
		"migrate",
	}
	allArgs = append(allArgs, defaultArgs...)
	cmd := commandExecutor("flyway", allArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("flyway migration failed: %w", err)
	}

	return nil
}
