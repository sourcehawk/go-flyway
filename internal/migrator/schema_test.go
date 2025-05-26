package migrator

import (
	"os/exec"
	"testing"

	cp "github.com/sourcehawk/go-flyway/internal/credentials_provider"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func validTestSchema() *Schema {
	return &Schema{
		Name:           "name",
		MigrationsPath: "path",
		FlywayArgs: []string{
			"-key1=val1",
			"-key2=val2",
		},
		Credentials: validTestCredentials(),
	}
}

func Test_Schema_Validate_Succeeds(t *testing.T) {
	s := validTestSchema()
	assert := assert.New(t)
	assert.NoError(s.Validate())
}

func Test_Schema_Validate_FailsOnMissingName(t *testing.T) {
	s := validTestSchema()
	s.Name = ""
	assert := assert.New(t)
	assert.Error(s.Validate())
}

func Test_Schema_Validate_FailsOnMissingMigrationsPath(t *testing.T) {
	s := validTestSchema()
	s.MigrationsPath = ""
	assert := assert.New(t)
	assert.Error(s.Validate())
}

func Test_Schema_Validate_FailsOnMissingCredentials(t *testing.T) {
	s := validTestSchema()
	s.Credentials = nil
	assert := assert.New(t)
	assert.Error(s.Validate())
}

func Test_Schema_Validate_FailsIfFetchingCredentialsFails(t *testing.T) {
	s := validTestSchema()
	s.Credentials.TextProviderImpl.Database = ""
	assert := assert.New(t)
	assert.Error(s.Validate())
}

func Test_Schema_Validate_FailsIfInvalidFlywayArg(t *testing.T) {
	s := validTestSchema()
	assert := assert.New(t)

	s.FlywayArgs = []string{"-invalid=because=twoequalsigns"}
	assert.Error(s.Validate())

	s.FlywayArgs = []string{"-invalidbecausenoequalssign"}
	assert.Error(s.Validate())

	s.FlywayArgs = []string{"invalidbecausenodash=value"}
	assert.Error(s.Validate())

	s.FlywayArgs = []string{"-="}
	assert.Error(s.Validate())

	s.FlywayArgs = []string{""}
	assert.Error(s.Validate())
}

func Test_Schema_Validate_FailsIfPlaceholderInvalid(t *testing.T) {
	s := validTestSchema()
	s.Placeholders = append(s.Placeholders, &Placeholder{})

	assert := assert.New(t)
	assert.Error(s.Validate())
}

func Test_Schema_Validate_SucessFromYaml(t *testing.T) {
	assert := assert.New(t)

	data := `
name: testing
migrationsPath: ./data
credentials:
  provider: text
  text:
    username: x
    password: x
    host: x
    port: 6543
    database: x
flywayArgs:
  - -baselineOnMigrate=true
placeholders:
  - name: test_placeholder
    value: test_value
`
	s := &Schema{}
	assert.NoError(yaml.Unmarshal([]byte(data), s))

	assert.Equal(s.Name, "testing")
	assert.Equal(s.FlywayArgs, []string{"-baselineOnMigrate=true"})
	assert.Equal(*s.Placeholders[0], Placeholder{Name: "test_placeholder", Value: "test_value"})
	assert.Equal(s.Credentials.TextProviderImpl.DatabaseCredentials, cp.DatabaseCredentials{
		Username: "x",
		Password: "x",
		Host:     "x",
		Port:     6543,
		Database: "x",
	})
}

func Test_Schema_SetDefaultFlywayArgs_PicksCurrentOverDefaults(t *testing.T) {
	s := validTestSchema()
	s.FlywayArgs = []string{"-k1=v1", "-k2=v2", "-k3=v3"}
	err := s.SetDefaultFlywayArgs([]string{"-k4=v4", "-k5=v5", "-k1=vX"})

	assert := assert.New(t)
	assert.NoError(err)
	assert.Contains(s.FlywayArgs, "-k1=v1", "Should not be overwritten")
	assert.Contains(s.FlywayArgs, "-k2=v2")
	assert.Contains(s.FlywayArgs, "-k3=v3")
	assert.Contains(s.FlywayArgs, "-k4=v4")
	assert.Contains(s.FlywayArgs, "-k5=v5")
}

func Test_Schema_SetDefaultFlywayArgs_FailsOnInvalidFlywayArg(t *testing.T) {
	s := validTestSchema()
	assert := assert.New(t)

	s.FlywayArgs = []string{"-k1=v1", "-k2=v2", "-k3=v3"}
	assert.Error(s.SetDefaultFlywayArgs([]string{"-k4=v4", "-k5=v5", "-k1==vX"}))

	s.FlywayArgs = []string{"-k1=v1", "-k2==v2", "-k3=v3"}
	assert.Error(s.SetDefaultFlywayArgs([]string{"-k4=v4", "-k5=v5", "-k1=vX"}))
}

func Test_Schema_ensureFlyway_ReturnsWithoutErrorWhenFlywayInstalled(t *testing.T) {
	s := validTestSchema()

	assert := assert.New(t)

	err := s.ensureFlyway(func(name string, arg ...string) *exec.Cmd {
		assert.Equal(name, "flyway")
		return exec.Command("echo", "testing")
	})

	assert.NoError(err)
}

func Test_Schema_ensureFlyway_ReturnsErrorWhenFlywayNotInstalled(t *testing.T) {
	s := validTestSchema()
	assert := assert.New(t)

	err := s.ensureFlyway(func(name string, arg ...string) *exec.Cmd {
		assert.Equal(name, "flyway")
		return exec.Command("exit", "1")
	})

	assert.Error(err)
}

func Test_Schema_Migrate_AppliesCorrectSettingsToCommandExec(t *testing.T) {
	s := Schema{
		Name:           "test",
		MigrationsPath: "./data",
		Credentials:    validTestCredentials(),
		FlywayArgs: []string{
			"-baselineOnMigrate=true",
		},
		Placeholders: []*Placeholder{
			{
				Name:  "test_placeholder",
				Value: "test_replacement",
			},
		},
	}
	assert := assert.New(t)
	callcount := 0
	var execCmd *exec.Cmd

	err := s.Migrate(func(name string, arg ...string) *exec.Cmd {
		assert.Equal(name, "flyway")

		if callcount > 0 {
			assert.Contains(arg, "-baselineOnMigrate=true")
			assert.Contains(arg, "-locations=filesystem:./data")
			assert.Contains(arg, "-schemas=test")
			assert.Contains(arg, "-user=a")
			assert.Contains(arg, "-password=a")
			assert.Contains(arg, "-url=jdbc:postgresql://a:5432/a")
			assert.Contains(arg, "-placeholders.test_placeholder=test_replacement")
		}
		execCmd = exec.Command("echo", "testing")
		callcount += 1
		return execCmd
	})

	assert.NoError(err)
}

func Test_Schema_Migrate_FailsOnValidationErrors(t *testing.T) {
	s := validTestSchema()
	s.Name = ""
	callCount := 0
	err := s.Migrate(func(name string, arg ...string) *exec.Cmd {
		callCount++
		return exec.Command("echo", "testing")
	})

	assert := assert.New(t)
	assert.Error(err)
	assert.Equal(callCount, 0)
}

func Test_Schema_Migrate_FailsOnFlywayNotInstalled(t *testing.T) {
	s := validTestSchema()
	callcount := 0

	err := s.Migrate(func(name string, arg ...string) *exec.Cmd {
		callcount++
		return exec.Command("exit", "1")
	})

	assert := assert.New(t)
	assert.Error(err)
	assert.Equal(callcount, 1)
}

func Test_Schema_Migrate_FailsOnMigrationCommandErrors(t *testing.T) {
	s := validTestSchema()
	callcount := 0

	err := s.Migrate(func(name string, arg ...string) *exec.Cmd {
		if callcount > 0 {
			return exec.Command("exit", "1")
		}
		callcount += 1
		return exec.Command("echo", "test")
	})

	assert := assert.New(t)
	assert.Error(err)
	assert.Greater(callcount, 0)
}
