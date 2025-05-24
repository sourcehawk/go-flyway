package credentials_provider

import (
	"encoding/json"
	"fmt"
	"testing"

	sp "github.com/sourcehawk/go-flyway/internal/secrets_provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSecretsProvider struct {
	mock.Mock
}

func (m *MockSecretsProvider) GetSecret(name string) (map[string]any, error) {
	args := m.Called(name)
	return args.Get(0).(map[string]any), args.Error(1)
}

func validAWSSMDatabaseCredentials() *AWSSMDatabaseCredentials {
	return &AWSSMDatabaseCredentials{
		Username: &sp.SecretRef{SecretName: "a", SecretKey: "b"},
		Password: &sp.SecretRef{SecretName: "a", SecretKey: "b"},
		Host:     &sp.SecretRef{SecretName: "a", SecretKey: "b"},
		Port:     &sp.SecretRef{SecretName: "a", SecretKey: "b"},
		Database: &sp.SecretRef{SecretName: "a", SecretKey: "b"},
		awssm: new(MockSecretsProvider),
	}
}

func Test_AWSSMDatabaseCredentials_Validate_Success(t *testing.T) {
	c := validAWSSMDatabaseCredentials()
	assert := assert.New(t)
	assert.NoError(c.Validate())
}

func Test_AWSSMDatabaseCredentials_Validate_FailsWhenSecretRefInvalid(t *testing.T) {
	c := *validAWSSMDatabaseCredentials()
	c.Database = &sp.SecretRef{SecretKey: "b"}
	
	assert := assert.New(t)
	assert.Error(c.Validate())
}

func Test_AWSSMDatabaseCredentials_Validate_FailsWhenFieldNil(t *testing.T) {
	c := validAWSSMDatabaseCredentials()
	assert := assert.New(t)
	for _, field := range []**sp.SecretRef{&c.Username, &c.Password, &c.Host, &c.Port, &c.Database} {
		fieldBefore := *field
		*field = nil
		assert.Error(c.Validate())
		*field = fieldBefore
	}
}

func Test_AWSSMDatabaseCredentials_Validate_LoadsAWSProvider(t *testing.T) {
	c := validAWSSMDatabaseCredentials()
	c.awssm = nil
	assert := assert.New(t)
	calls := 0
	NewAWSSecretsManager = func() (*sp.AWSSecretsManager, error) {
		calls ++
		return &sp.AWSSecretsManager{}, nil
	}
	assert.NoError(c.Validate())
	assert.Equal(calls, 1)
}

func Test_AWSSMDatabaseCredentials_Validate_FailsWhenAWSProviderLoadFails(t *testing.T) {
	c := validAWSSMDatabaseCredentials()
	c.awssm = nil
	assert := assert.New(t)
	calls := 0
	NewAWSSecretsManager = func() (*sp.AWSSecretsManager, error) {
		calls ++
		return nil, fmt.Errorf("error")
	}
	assert.Error(c.Validate())
	assert.Equal(calls, 1)
}

func Test_AWSSMDatabaseCredentials_GetCredentials_Succeeds(t *testing.T) {
	awssm := new(MockSecretsProvider)
	fakeSecret := map[string]any{
		"usernamey": "bob",
		"passwordy": "supersecret",
		"hosty":     "localhost",
		"porty":     5432,
		"databasey": "postgres",
	}
	awssm.On("GetSecret", "a").Return(fakeSecret, nil)

	c := &AWSSMDatabaseCredentials{
		Username: &sp.SecretRef{SecretName: "a", SecretKey: "usernamey"},
		Password: &sp.SecretRef{SecretName: "a", SecretKey: "passwordy"},
		Host:     &sp.SecretRef{SecretName: "a", SecretKey: "hosty"},
		Port:     &sp.SecretRef{SecretName: "a", SecretKey: "porty"},
		Database: &sp.SecretRef{SecretName: "a", SecretKey: "databasey"},
		awssm:    awssm,
	}

	creds, err := c.GetCredentials()
	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(creds.Username, "bob")
	assert.Equal(creds.Password, "supersecret")
	assert.Equal(creds.Host, "localhost")
	assert.Equal(creds.Port, 5432)
	assert.Equal(creds.Database, "postgres")
}

type TestNotJSONSerializable struct {
	Ch chan struct{}
}

func Test_AWSSMDatabaseCredentials_GetCredentials_FailsWhenSecretValueNotJsonSerializable(t *testing.T) {
	awssm := new(MockSecretsProvider)
	fakeSecret := map[string]any{
		"usernamey": TestNotJSONSerializable{},
		"passwordy": "supersecret",
		"hosty":     "localhost",
		"porty":     5432,
		"databasey": "postgres",
	}
	awssm.On("GetSecret", "a").Return(fakeSecret, nil)

	c := &AWSSMDatabaseCredentials{
		Username: &sp.SecretRef{SecretName: "a", SecretKey: "usernamey"},
		Password: &sp.SecretRef{SecretName: "a", SecretKey: "passwordy"},
		Host:     &sp.SecretRef{SecretName: "a", SecretKey: "hosty"},
		Port:     &sp.SecretRef{SecretName: "a", SecretKey: "porty"},
		Database: &sp.SecretRef{SecretName: "a", SecretKey: "databasey"},
		awssm:    awssm,
	}

	_, err := c.GetCredentials()
	assert := assert.New(t)
	var targetErr *json.UnsupportedTypeError
	assert.Error(err)
	assert.ErrorAs(err, &targetErr)
}

func Test_AWSSMDatabaseCredentials_GetCredentials_FailsWhenJsonNotParsableToStruct(t *testing.T) {
	awssm := new(MockSecretsProvider)
	fakeSecret := map[string]any{
		"usernamey": 10,
		"passwordy": "supersecret",
		"hosty":     "localhost",
		"porty":     5432,
		"databasey": "postgres",
	}
	awssm.On("GetSecret", "a").Return(fakeSecret, nil)

	c := &AWSSMDatabaseCredentials{
		Username: &sp.SecretRef{SecretName: "a", SecretKey: "usernamey"},
		Password: &sp.SecretRef{SecretName: "a", SecretKey: "passwordy"},
		Host:     &sp.SecretRef{SecretName: "a", SecretKey: "hosty"},
		Port:     &sp.SecretRef{SecretName: "a", SecretKey: "porty"},
		Database: &sp.SecretRef{SecretName: "a", SecretKey: "databasey"},
		awssm:    awssm,
	}

	_, err := c.GetCredentials()
	assert := assert.New(t)
	var targetErr *json.UnmarshalTypeError
	assert.Error(err)
	assert.ErrorAs(err, &targetErr)
}

func Test_AWSSMDatabaseCredentials_GetCredentials_FailsWhenValidationError(t *testing.T) {
	c := &AWSSMDatabaseCredentials{
		Username: &sp.SecretRef{SecretName: "a", SecretKey: "usernamey"},
		Password: &sp.SecretRef{SecretName: "a", SecretKey: "passwordy"},
		Host:     &sp.SecretRef{SecretName: "a", SecretKey: "hosty"},
		Port:     &sp.SecretRef{SecretName: "a", SecretKey: "porty"},
		Database: nil,
		awssm:    new(MockSecretsProvider),
	}
	_, err := c.GetCredentials()
	assert := assert.New(t)
	assert.Error(err)
}

func Test_AWSSMDatabaseCredentials_GetCredentials_FailsOnSecretsProviderError(t *testing.T) {
	awssm := new(MockSecretsProvider)

	c := &AWSSMDatabaseCredentials{
		Username: &sp.SecretRef{SecretName: "a", SecretKey: "usernamey"},
		Password: &sp.SecretRef{SecretName: "a", SecretKey: "passwordy"},
		Host:     &sp.SecretRef{SecretName: "a", SecretKey: "hosty"},
		Port:     &sp.SecretRef{SecretName: "a", SecretKey: "porty"},
		Database: &sp.SecretRef{SecretName: "a", SecretKey: "databasey"},
		awssm:    awssm,
	}

	fakeSecret := make(map[string]any)
	awssm.On("GetSecret", mock.Anything).Return(fakeSecret, fmt.Errorf("test error 123"))

	_, err := c.GetCredentials()
	awssm.AssertCalled(t, "GetSecret", "a")
	assert := assert.New(t)
	assert.Error(err)
}
