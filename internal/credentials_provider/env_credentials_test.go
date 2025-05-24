package credentials_provider

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func validEnvCredentials(t *testing.T) *EnvDatabaseCredentials {
	e := &EnvDatabaseCredentials{
		UsernameKey: "UsernameEnvKey",
		PasswordKey: "PasswordEnvKey",
		HostKey:     "HostEnvKey",
		PortKey:     "PortEnvKey",
		DatabaseKey: "DatabasEnvKey",
	}
	t.Setenv(e.UsernameKey, "user")
	t.Setenv(e.PasswordKey, "pass")
	t.Setenv(e.HostKey, "host")
	t.Setenv(e.PortKey, "5432")
	t.Setenv(e.DatabaseKey, "database")

	return e
}

func Test_EnvDatabaseCredentials_Validate_Succeeds(t *testing.T) {
	e := validEnvCredentials(t)
	assert := assert.New(t)
	assert.NoError(e.Validate())
}

func Test_EnvDatabaseCredentials_Validate_FailsWhenPortNotInteger(t *testing.T) {
	assert := assert.New(t)

	e := validEnvCredentials(t)
	t.Setenv(e.PortKey, "notint")
	assert.Error(e.Validate())
}

func Test_EnvDatabaseCredentials_Validate_FailsWhenMissingCredentialsField(t *testing.T) {
	assert := assert.New(t)

	e := validEnvCredentials(t)
	e.UsernameKey = ""
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	e.PasswordKey = ""
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	e.HostKey = ""
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	e.PortKey = ""
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	e.DatabaseKey = ""
	assert.Error(e.Validate())
}

func Test_EnvDatabaseCredentials_Validate_FailsWhenMissingEnvKey(t *testing.T) {
	assert := assert.New(t)

	e := validEnvCredentials(t)
	os.Unsetenv(e.UsernameKey) //nolint:errcheck
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	os.Unsetenv(e.PasswordKey) //nolint:errcheck
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	os.Unsetenv(e.HostKey) //nolint:errcheck
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	os.Unsetenv(e.PortKey) //nolint:errcheck
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	os.Unsetenv(e.DatabaseKey) //nolint:errcheck
	assert.Error(e.Validate())
}

func Test_EnvDatabaseCredentials_Validate_FailsWhenEnvValueEmpty(t *testing.T) {
	assert := assert.New(t)

	e := validEnvCredentials(t)
	t.Setenv(e.UsernameKey, "")
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	t.Setenv(e.PasswordKey, "")
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	t.Setenv(e.HostKey, "")
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	t.Setenv(e.PortKey, "")
	assert.Error(e.Validate())

	e = validEnvCredentials(t)
	t.Setenv(e.DatabaseKey, "")
	assert.Error(e.Validate())
}

func Test_EnvDatabaseCredentials_GetCredentials_Succeeds(t *testing.T) {
	e := validEnvCredentials(t)
	creds, err := e.GetCredentials()

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(*creds, DatabaseCredentials{
		Username: "user",
		Password: "pass",
		Host:     "host",
		Port:     5432,
		Database: "database",
	})
}

func Test_EnvDatabaseCredentials_GetCredentials_FailsOnValidationError(t *testing.T) {
	e := validEnvCredentials(t)
	e.DatabaseKey = ""
	_, err := e.GetCredentials()

	assert := assert.New(t)
	assert.Error(err)
}
