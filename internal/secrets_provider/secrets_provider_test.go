package secrets_provider

import (
	"encoding/json"
	"fmt"
	"testing"

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

type TestSecrets struct {
	Foo string  `json:"foo,omitempty"`
	Bar int     `json:"bar,omitempty"`
	Baz float32 `json:"baz,omitempty"`
}

func Test_SecretRef_Validate_Succeeds(t *testing.T) {
	assert := assert.New(t)
	secretRef := &SecretRef{SecretName: "foo", SecretKey: "bar"}
	assert.NoError(secretRef.Validate())
}

func Test_SecretRef_Validate_FailsWithMissingSecretName(t *testing.T) {
	assert := assert.New(t)
	secretRef := &SecretRef{SecretKey: "bar"}
	assert.Error(secretRef.Validate())
}

func Test_SecretRef_Validate_FailsWithMissingSecretKey(t *testing.T) {
	assert := assert.New(t)
	secretRef := &SecretRef{SecretName: "foo"}
	assert.Error(secretRef.Validate())
}

func Test_SecretRefToStructJsonField_WorksAndCachesSecrets(t *testing.T) {
	sp := new(MockSecretsProvider)
	spSecret := map[string]any{"1": "hello", "2": float64(45), "3": float64(13.4)}
	sp.On("GetSecret", "test").Return(spSecret, nil)

	secrets := make(map[string]map[string]any)
	jsonStruct := make(map[string]any)
	assert := assert.New(t)

	for _, s := range []SecretRefToStructJsonField{
		{StructJsonField: "foo", SecretRef: &SecretRef{SecretName: "test", SecretKey: "1"}},
		{StructJsonField: "bar", SecretRef: &SecretRef{SecretName: "test", SecretKey: "2"}},
		{StructJsonField: "baz", SecretRef: &SecretRef{SecretName: "test", SecretKey: "3"}},
	} {
		err := s.PopulateJSONFieldFromSecret(sp, secrets, jsonStruct)
		assert.NoError(err)
	}

	sp.AssertNumberOfCalls(t, "GetSecret", 1)
	assert.Contains(jsonStruct, "foo")
	assert.Contains(jsonStruct, "bar")
	assert.Contains(jsonStruct, "baz")
	assert.Equal(jsonStruct["foo"], "hello")
	assert.Equal(jsonStruct["bar"], float64(45))
	assert.Equal(jsonStruct["baz"], float64(13.4))

	data, err := json.Marshal(jsonStruct)
	assert.Nil(err)
	concreteStruct := &TestSecrets{}
	err = json.Unmarshal(data, concreteStruct)
	assert.Nil(err)
	assert.Equal(concreteStruct.Foo, "hello")
	assert.Equal(concreteStruct.Bar, 45)
	assert.Equal(concreteStruct.Baz, float32(13.4))
}

func Test_SecretRefToStructJsonField_FailsWhenSecretMissing(t *testing.T) {
	sp := new(MockSecretsProvider)
	spSecret := map[string]any{}
	sp.On("GetSecret", "test").Return(spSecret, fmt.Errorf("not found"))

	secrets := make(map[string]map[string]any)
	jsonStruct := make(map[string]any)

	secretRefJson := SecretRefToStructJsonField{
		StructJsonField: "foo",
		SecretRef:       &SecretRef{SecretName: "test", SecretKey: "NOT_CORRECT_KEY"},
	}
	err := secretRefJson.PopulateJSONFieldFromSecret(sp, secrets, jsonStruct)

	assert := assert.New(t)
	assert.Error(err)
	assert.NotContains(secrets, "test")
	sp.AssertNumberOfCalls(t, "GetSecret", 1)
}

func Test_SecretRefToStructJsonField_FailsWhenSecretKeyMissing(t *testing.T) {
	sp := new(MockSecretsProvider)
	spSecret := map[string]any{"1": "hello"}
	sp.On("GetSecret", "test").Return(spSecret, nil)

	secrets := make(map[string]map[string]any)
	jsonStruct := make(map[string]any)

	secretRefJson := SecretRefToStructJsonField{
		StructJsonField: "foo",
		SecretRef:       &SecretRef{SecretName: "test", SecretKey: "NOT_CORRECT_KEY"},
	}
	err := secretRefJson.PopulateJSONFieldFromSecret(sp, secrets, jsonStruct)

	assert := assert.New(t)
	assert.Error(err)
	assert.NotContains(jsonStruct, "foo")
	sp.AssertNumberOfCalls(t, "GetSecret", 1)
}

func Test_SecretRefToStructJsonField_FailsWhenSecretValueNil(t *testing.T) {
	sp := new(MockSecretsProvider)
	spSecret := map[string]any{"1": nil}
	sp.On("GetSecret", "test").Return(spSecret, nil)

	secrets := make(map[string]map[string]any)
	jsonStruct := make(map[string]any)

	secretRefJson := SecretRefToStructJsonField{
		StructJsonField: "foo",
		SecretRef:       &SecretRef{SecretName: "test", SecretKey: "1"},
	}
	err := secretRefJson.PopulateJSONFieldFromSecret(sp, secrets, jsonStruct)

	assert := assert.New(t)
	assert.Error(err)
	assert.NotContains(jsonStruct, "foo")
	sp.AssertNumberOfCalls(t, "GetSecret", 1)
}

func Test_SecretRefToStructJsonField_PanicsWhenSecretsUninitialized(t *testing.T) {
	sp := new(MockSecretsProvider)

	var secrets map[string]map[string]any

	jsonStruct := make(map[string]any)
	secret := SecretRefToStructJsonField{
		StructJsonField: "foo",
		SecretRef:       &SecretRef{SecretName: "test", SecretKey: "1"},
	}

	assert := assert.New(t)
	assert.Panics(func() {
		secret.PopulateJSONFieldFromSecret(sp, secrets, jsonStruct) //nolint:errcheck
	})
}

func Test_SecretRefToStructJsonField_PanicsWhenJsonStructUninitialized(t *testing.T) {
	sp := new(MockSecretsProvider)

	secrets := make(map[string]map[string]any)
	var jsonStruct map[string]any

	secret := SecretRefToStructJsonField{
		StructJsonField: "foo",
		SecretRef:       &SecretRef{SecretName: "test", SecretKey: "1"},
	}

	assert := assert.New(t)
	assert.Panics(func() {
		secret.PopulateJSONFieldFromSecret(sp, secrets, jsonStruct) //nolint:errcheck
	})
}

func Test_SecretRefToStructJsonField_FailsWhenSecretRefNil(t *testing.T) {
	sp := new(MockSecretsProvider)

	secrets := make(map[string]map[string]any)
	jsonStruct := make(map[string]any)

	secret := SecretRefToStructJsonField{
		StructJsonField: "foo",
		SecretRef:       nil,
	}

	assert := assert.New(t)
	assert.Error(secret.PopulateJSONFieldFromSecret(sp, secrets, jsonStruct))
}
