package migrator

import (
	"fmt"
	"testing"

	cp "github.com/sourcehawk/go-flyway/internal/credentials_provider"
	sp "github.com/sourcehawk/go-flyway/internal/secrets_provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v3"
)

type MockCredentialsProvider struct {
	mock.Mock
}

// Valid credentials object with all fields set to "a" and port set to 5432
func validTestCredentials() *Credentials {
	return &Credentials{
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
	}
}

func (m *MockCredentialsProvider) GetCredentials() (*cp.DatabaseCredentials, error) {
	args := m.Called()
	return args.Get(0).(*cp.DatabaseCredentials), args.Error(1)
}

func (m *MockCredentialsProvider) Validate() error {
	args := m.Called()
	return args.Error(0)
}

func Test_Credentials_Validate_Succeeds(t *testing.T) {
	c := validTestCredentials()
	assert := assert.New(t)
	assert.NoError(c.Validate())
}

func Test_Credentials_Validate_FailsIfNoProviderSpecified(t *testing.T) {
	c := validTestCredentials()
	c.Provider = ""
	assert := assert.New(t)
	assert.Error(c.Validate())
}

func Test_Credentials_Validate_FailsIfInvalidProviderSpecified(t *testing.T) {
	c := validTestCredentials()
	c.Provider = "invalid"
	assert := assert.New(t)
	assert.Error(c.Validate())
}

func Test_Credentials_Validate_EnvDatabaseCredentials(t *testing.T) {
	envCreds := &cp.EnvDatabaseCredentials{
		UsernameKey: "UserEnvKey",
		PasswordKey: "PasswordEnvKey",
		HostKey:     "HostEnvKey",
		PortKey:     "PortEnvKey",
		DatabaseKey: "DatabaseEnvKey",
	}

	t.Setenv(envCreds.UsernameKey, "user")
	t.Setenv(envCreds.PasswordKey, "pass")
	t.Setenv(envCreds.HostKey, "host")
	t.Setenv(envCreds.PortKey, "5432")
	t.Setenv(envCreds.DatabaseKey, "database")

	c := Credentials{
		Provider: string(cp.EnvProviderType),
		CredentialProviders: CredentialProviders{
			EnvProviderImpl: envCreds,
		},
	}
	assert := assert.New(t)
	assert.NoError(c.Validate())
}

func Test_Credentials_Validate_EnvDatabaseCredentialsFailsIfNoImpl(t *testing.T) {
	c := Credentials{
		Provider:            string(cp.EnvProviderType),
		CredentialProviders: CredentialProviders{},
	}
	assert := assert.New(t)
	assert.Error(c.Validate())
}

func Test_Credentials_Validate_EnvDatabaseCredentialsFromYaml(t *testing.T) {
	t.Setenv("USER_KEY", "user")
	t.Setenv("PASSWORD_KEY", "pass")
	t.Setenv("HOST_KEY", "host")
	t.Setenv("PORT_KEY", "5432")
	t.Setenv("DATABASE_KEY", "database")

	data := `
provider: env
env:
  usernameKey: USER_KEY
  passwordKey: PASSWORD_KEY
  hostKey: HOST_KEY
  portKey: PORT_KEY
  databaseKey: DATABASE_KEY
`

	c := &Credentials{}
	assert := assert.New(t)
	assert.NoError(yaml.Unmarshal([]byte(data), c))
	assert.NoError(c.Validate())
}

func Test_Credentials_Validate_AWSSMDatabaseCredentials(t *testing.T) {
	c := Credentials{
		Provider: string(cp.AWSSMSecretsProviderType),
		CredentialProviders: CredentialProviders{
			AwssmProviderImpl: &cp.AWSSMDatabaseCredentials{
				Username: &sp.SecretRef{SecretName: "a", SecretKey: "b"},
				Password: &sp.SecretRef{SecretName: "a", SecretKey: "b"},
				Host:     &sp.SecretRef{SecretName: "a", SecretKey: "b"},
				Port:     &sp.SecretRef{SecretName: "a", SecretKey: "b"},
				Database: &sp.SecretRef{SecretName: "a", SecretKey: "b"},
			},
		},
	}
	assert := assert.New(t)
	assert.NoError(c.Validate())
}

func Test_Credentials_Validate_AWSSMDatabaseCredentialsFailsIfNoImpl(t *testing.T) {
	c := Credentials{
		Provider:            string(cp.AWSSMSecretsProviderType),
		CredentialProviders: CredentialProviders{},
	}
	assert := assert.New(t)
	assert.Error(c.Validate())
}

func Test_Credentials_Validate_AWSSMDatabaseCredentialsFromYaml(t *testing.T) {
	data := `
provider: aws_sm
aws_sm:
  username:
    secretName: a
    secretKey: b
  password:
    secretName: a
    secretKey: b
  host:
    secretName: a
    secretKey: b
  port:
    secretName: a
    secretKey: b
  database:
    secretName: a
    secretKey: b
`

	c := &Credentials{}
	assert := assert.New(t)
	assert.NoError(yaml.Unmarshal([]byte(data), c))
	assert.NoError(c.Validate())
}

func Test_Credentials_Validate_TextDatabaseCredentials(t *testing.T) {
	c := Credentials{
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
	}
	assert := assert.New(t)
	assert.NoError(c.Validate())
}

func Test_Credentials_Validate_TextDatabaseCredentialsFailsIfNoImpl(t *testing.T) {
	c := Credentials{
		Provider:            string(cp.TextProviderType),
		CredentialProviders: CredentialProviders{},
	}
	assert := assert.New(t)
	assert.Error(c.Validate())
}

func Test_Credentials_Validate_TextDatabaseCredentialsFromYaml(t *testing.T) {
	data := `
provider: text
text:
  username: a
  password: a
  host: a
  port: 5432
  database: a
`

	c := &Credentials{}
	assert := assert.New(t)
	assert.NoError(yaml.Unmarshal([]byte(data), c))
	assert.NoError(c.Validate())
}

func Test_Credentials_fetchCredentials_SucceedsAndCachesFetchedCredentials(t *testing.T) {
	c := Credentials{
		Provider: string(cp.TextProviderType),
		CredentialProviders: CredentialProviders{
			TextProviderImpl: &cp.DatabaseCredentials{
				Username: "a",
				Password: "b",
				Host:     "c",
				Port:     5432,
				Database: "d",
			},
		},
	}
	assert := assert.New(t)
	err := c.Validate()
	assert.NoError(err)

	creds, err := c.fetchCredentials()
	assert.NoError(err)
	assert.Equal(creds, c.TextProviderImpl)

	mockConcreteProvider := new(MockCredentialsProvider)
	mockConcreteProvider.On("GetCredentials").Return(&cp.DatabaseCredentials{}, nil)
	mockConcreteProvider.On("Validate").Return(nil)

	c.concreteProvider = mockConcreteProvider
	creds, err = c.fetchCredentials()
	mockConcreteProvider.AssertNotCalled(t, "GetCredentials")
	assert.NoError(err)
	assert.Equal(creds, c.TextProviderImpl)
}

func Test_Credentials_fetchCredentials_FailsWhenCredentialProviderReturnsError(t *testing.T) {
	c := Credentials{
		Provider: string(cp.TextProviderType),
		CredentialProviders: CredentialProviders{
			TextProviderImpl: &cp.DatabaseCredentials{
				Username: "a",
				Password: "b",
				Host:     "c",
				Port:     5432,
				Database: "d",
			},
		},
	}

	assert := assert.New(t)
	err := c.Validate()
	assert.NoError(err)

	mockConcreteProvider := new(MockCredentialsProvider)
	mockConcreteProvider.On("GetCredentials").Return(&cp.DatabaseCredentials{}, fmt.Errorf("test error 123"))
	mockConcreteProvider.On("Validate").Return(nil)
	c.concreteProvider = mockConcreteProvider
	_, err = c.fetchCredentials()
	mockConcreteProvider.AssertCalled(t, "GetCredentials")
	assert.Error(err)
}

func Test_Credentials_FetchCredentials_Succeeds(t *testing.T) {
	c := Credentials{
		Provider: string(cp.TextProviderType),
		CredentialProviders: CredentialProviders{
			TextProviderImpl: &cp.DatabaseCredentials{
				Username: "a",
				Password: "b",
				Host:     "c",
				Port:     5432,
				Database: "d",
			},
		},
	}
	assert := assert.New(t)
	creds, err := c.FetchCredentials()
	assert.NoError(err)
	assert.Equal(creds, c.TextProviderImpl)
}

func Test_Credentials_FetchCredentials_FailsOnValidationError(t *testing.T) {
	c := Credentials{
		Provider: string(cp.TextProviderType),
		CredentialProviders: CredentialProviders{
			TextProviderImpl: &cp.DatabaseCredentials{
				Username: "a",
				Password: "b",
				Host:     "c",
				Port:     5432,
				// Database: "d",
			},
		},
	}
	assert := assert.New(t)
	_, err := c.FetchCredentials()
	assert.Error(err)
}
