package migrator

import (
	"os"
	"os/exec"
	"testing"

	cp "github.com/sourcehawk/go-flyway/internal/credentials_provider"
	"github.com/stretchr/testify/assert"
)

func validMockMigrator() *Migrator {
	return &Migrator{
		FlywayArgs: []string{"-mykey=myvalue"},
		Credentials: &Credentials{
			Provider: string(cp.TextProviderType),
			CredentialProviders: CredentialProviders{
				TextProviderImpl: &cp.DatabaseCredentials{
					Username: "a",
					Password: "a",
					Host:     "a",
					Port:     5432,
					Database: "a",
				},
			},
		},
		Schemas: []*Schema{
			{
				Name:           "foo",
				MigrationsPath: "./data/foo",
				Placeholders:   []*Placeholder{{Name: "p1", Value: "v1"}}},
			{
				Name:           "bar",
				MigrationsPath: "./data/bar",
			},
		},
		cmdExecFunc: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("echo", "testing")
		},
	}
}

func Test_Migrator_Validate_Succeeds(t *testing.T) {
	m := validMockMigrator()
	assert := assert.New(t)
	assert.NoError(m.Validate())
}

func Test_Migrator_Validate_FailsWhenInvalidSchema(t *testing.T) {
	m := validMockMigrator()
	m.Schemas = append(m.Schemas, &Schema{})
	assert := assert.New(t)
	assert.Error(m.Validate())
}

func Test_Migrator_Validate_FailsWhenInvalidDefaultCredentials(t *testing.T) {
	m := validMockMigrator()
	m.Credentials.TextProviderImpl.Database = ""
	assert := assert.New(t)
	assert.Error(m.Validate())
}

func Test_Migrator_Validate_FailsWhenInvalidDefaultFlywayArgs(t *testing.T) {
	m := validMockMigrator()
	m.FlywayArgs = []string{"invalid"}
	assert := assert.New(t)
	assert.Error(m.Validate())
}

func Test_Migrator_Validate_FailsWhenNoDefaultCredentialsAndNoSchemaCredentials(t *testing.T) {
	m := validMockMigrator()
	m.Credentials = nil
	m.Schemas = []*Schema{validTestSchema()}
	m.Schemas[0].Credentials = nil

	assert := assert.New(t)
	assert.Error(m.Validate())
}

func Test_Migrator_Migrate_SucceedsWithValidConfig(t *testing.T) {
	m := validMockMigrator()
	assert := assert.New(t)
	assert.NoError(m.Migrate())
}

func Test_Migrator_Migrate_FailsWhenValidationError(t *testing.T) {
	m := validMockMigrator()
	m.Schemas[1] = &Schema{}
	m.cmdExecFunc = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "testing")
	}

	assert := assert.New(t)
	assert.Error(m.Migrate())
}

func Test_Migrator_Migrate_FailsWhenMigrationError(t *testing.T) {
	m := validMockMigrator()
	m.Schemas = append(m.Schemas, validTestSchema())
	assert := assert.New(t)
	assert.GreaterOrEqual(len(m.Schemas), 3, "Need at least 3 schemas for test")

	callCount := 0
	m.cmdExecFunc = func(name string, arg ...string) *exec.Cmd {
		if callCount >= 1 {
			return exec.Command("exit", "1")
		}
		callCount++
		return exec.Command("echo", "testing")
	}

	assert.Error(m.Migrate())
	assert.Equal(1, callCount, "Does not continue executing migrations after error")
}

func Test_NewMigrator_loadsFromConfigFileWithoutErrors(t *testing.T) {
	data := `
flywayArgs:
  - "-key=val"
credentials:
  provider: text
  text:
    username: x
    password: x
    host: x
    port: 5432
    database: x
schemas:
  - name: schema_1
    migrationsPath: ./data/schema_1
    flywayArgs:
      - "-key=val2"
      - "-foo=bar"
    credentials:
      provider: text
      text:
        username: y
        password: y
        host: y
        port: 6543
        database: y
  - name: schema_2
    migrationsPath: ./data/schema_2
    flywayArgs:
      - "-baselineOnMigrate=true"
    placeholders:
      - name: test_placeholder
        value: test_value
`
	path, err := writeTestFile("testfile", []byte(data))
	defer os.Remove(path) //nolint:errcheck

	assert := assert.New(t)
	assert.NoError(err)

	m, err := NewMigrator(path)
	assert.NoError(err)

	schema1 := m.Schemas[0]
	assert.Equal(schema1.Name, "schema_1")
	assert.Equal(schema1.MigrationsPath, "./data/schema_1")
	assert.Nil(schema1.Placeholders)
	assert.Contains(schema1.FlywayArgs, "-key=val2")
	assert.Contains(schema1.FlywayArgs, "-foo=bar")
	assert.Equal(*schema1.Credentials.TextProviderImpl, cp.DatabaseCredentials{
		Username: "y",
		Password: "y",
		Host:     "y",
		Port:     6543,
		Database: "y",
	})

	schema2 := m.Schemas[1]
	assert.Equal(schema2.Name, "schema_2")
	assert.Contains(schema2.FlywayArgs, "-key=val")
	assert.Contains(schema2.FlywayArgs, "-baselineOnMigrate=true")
	assert.Equal(*schema2.Placeholders[0], Placeholder{Name: "test_placeholder", Value: "test_value"})
	assert.Equal(*schema2.Credentials.TextProviderImpl, cp.DatabaseCredentials{
		Username: "x",
		Password: "x",
		Host:     "x",
		Port:     5432,
		Database: "x",
	})

}

func Test_NewMigrator_FailsWhenConfigFileNotFound(t *testing.T) {
	assert := assert.New(t)
	_, err := NewMigrator("path_that_doesnt_exist")
	assert.Error(err)
}

func Test_NewMigrator_FailsWhenConfigFileNotYamlParsable(t *testing.T) {
	path, err := writeTestFile("testfile", []byte("invalidyaml"))
	defer os.Remove(path) //nolint:errcheck

	assert := assert.New(t)
	assert.NoError(err)
	_, err = NewMigrator(path)
	assert.Error(err)
}
