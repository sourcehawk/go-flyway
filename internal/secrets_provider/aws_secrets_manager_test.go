package secrets_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAWSSecretsManagerClient struct {
	mock.Mock
}

func (m *MockAWSSecretsManagerClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*secretsmanager.GetSecretValueOutput), args.Error(1)
}

func Test_AWSSecretsManager_GetSecret_FetchesSecretFromString(t *testing.T) {
	awssmClient := new(MockAWSSecretsManagerClient)
	secretName := "foo/bar/baz"

	now := time.Now()
	secret := map[string]any{"hello": "world"}
	bytes, err := json.Marshal(secret)
	assert := assert.New(t)
	assert.Nil(err)

	output := &secretsmanager.GetSecretValueOutput{
		ARN:          aws.String(fmt.Sprintf("arn:aws:secretsmanager:eu-west-1:123456789:secret:%s", secretName)),
		CreatedDate:  &now,
		Name:         aws.String(secretName),
		SecretBinary: nil,
		SecretString: aws.String(string(bytes)),
	}

	awssmClient.On(
		"GetSecretValue",
		mock.Anything,
		mock.MatchedBy(func(p *secretsmanager.GetSecretValueInput) bool {
			return *p.SecretId == secretName
		}),
		mock.Anything,
	).Return(output, nil)

	awssm := AWSSecretsManager{
		client: awssmClient,
	}

	secretOut, err := awssm.GetSecret(secretName)
	assert.Nil(err)
	assert.Equal(secretOut, secret)
}

func Test_AWSSecretsManager_GetSecret_FetchesSecretFromBinary(t *testing.T) {
	awssmClient := new(MockAWSSecretsManagerClient)
	secretName := "foo/bar/baz"

	now := time.Now()
	secret := map[string]any{"hello": "world"}
	bytes, err := json.Marshal(secret)
	assert := assert.New(t)
	assert.Nil(err)

	output := &secretsmanager.GetSecretValueOutput{
		ARN:          aws.String(fmt.Sprintf("arn:aws:secretsmanager:eu-west-1:123456789:secret:%s", secretName)),
		CreatedDate:  &now,
		Name:         aws.String(secretName),
		SecretBinary: bytes,
		SecretString: nil,
	}

	awssmClient.On(
		"GetSecretValue",
		mock.Anything,
		mock.MatchedBy(func(p *secretsmanager.GetSecretValueInput) bool {
			return *p.SecretId == secretName
		}),
		mock.Anything,
	).Return(output, nil)

	awssm := AWSSecretsManager{
		client: awssmClient,
	}

	secretOut, err := awssm.GetSecret(secretName)
	assert.Nil(err)
	assert.Equal(secretOut, secret)
}

func Test_AWSSecretsManager_GetSecret_ErrorWhenNoStringOrBinary(t *testing.T) {
	awssmClient := new(MockAWSSecretsManagerClient)
	secretName := "foo/bar/baz"
	assert := assert.New(t)

	output := &secretsmanager.GetSecretValueOutput{}

	awssmClient.On(
		"GetSecretValue",
		mock.Anything,
		mock.MatchedBy(func(p *secretsmanager.GetSecretValueInput) bool {
			return *p.SecretId == secretName
		}),
		mock.Anything,
	).Return(output, nil)

	awssm := AWSSecretsManager{
		client: awssmClient,
	}

	_, err := awssm.GetSecret(secretName)
	assert.NotNil(err)
}

func Test_AWSSecretsManager_GetSecret_ErrorWhenGetSecretValueFailsWithAPIError(t *testing.T) {
	awssmClient := new(MockAWSSecretsManagerClient)
	assert := assert.New(t)

	errorOutput := &smtypes.ResourceNotFoundException{Message: aws.String("test")}

	awssmClient.On(
		"GetSecretValue",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(&secretsmanager.GetSecretValueOutput{}, errorOutput)

	awssm := AWSSecretsManager{
		client: awssmClient,
	}

	_, err := awssm.GetSecret("foo")
	assert.ErrorIs(err, errorOutput)
}

func Test_AWSSecretsManager_GetSecret_ErrorWhenContextTimeout(t *testing.T) {
	awssmClient := new(MockAWSSecretsManagerClient)
	assert := assert.New(t)

	awssmClient.On(
		"GetSecretValue",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(&secretsmanager.GetSecretValueOutput{}, context.DeadlineExceeded)

	awssm := AWSSecretsManager{
		client: awssmClient,
	}

	_, err := awssm.GetSecret("foo")
	t.Log(err.Error())
	assert.ErrorIs(err, context.DeadlineExceeded)
}

func Test_AWSSecretsManager_GetSecret_ErrorWhenSecretStringNotJsonParsable(t *testing.T) {
	awssmClient := new(MockAWSSecretsManagerClient)
	secretName := "foo/bar/baz"

	now := time.Now()
	assert := assert.New(t)

	output := &secretsmanager.GetSecretValueOutput{
		ARN:          aws.String(fmt.Sprintf("arn:aws:secretsmanager:eu-west-1:123456789:secret:%s", secretName)),
		CreatedDate:  &now,
		Name:         aws.String(secretName),
		SecretBinary: nil,
		SecretString: aws.String("notjson"),
	}

	awssmClient.On(
		"GetSecretValue",
		mock.Anything,
		mock.MatchedBy(func(p *secretsmanager.GetSecretValueInput) bool {
			return *p.SecretId == secretName
		}),
		mock.Anything,
	).Return(output, nil)

	awssm := AWSSecretsManager{
		client: awssmClient,
	}

	_, err := awssm.GetSecret(secretName)
	var syntaxError *json.SyntaxError
	assert.ErrorAs(err, &syntaxError, "error should be a syntax error")
}

func Test_AWSSecretsManager_GetSecret_ErrorWhenBinaryNotJsonParsable(t *testing.T) {
	awssmClient := new(MockAWSSecretsManagerClient)
	secretName := "foo/bar/baz"

	now := time.Now()
	bytes := []byte("notjson")
	assert := assert.New(t)

	output := &secretsmanager.GetSecretValueOutput{
		ARN:          aws.String(fmt.Sprintf("arn:aws:secretsmanager:eu-west-1:123456789:secret:%s", secretName)),
		CreatedDate:  &now,
		Name:         aws.String(secretName),
		SecretBinary: bytes,
		SecretString: nil,
	}

	awssmClient.On(
		"GetSecretValue",
		mock.Anything,
		mock.MatchedBy(func(p *secretsmanager.GetSecretValueInput) bool {
			return *p.SecretId == secretName
		}),
		mock.Anything,
	).Return(output, nil)

	awssm := AWSSecretsManager{
		client: awssmClient,
	}

	_, err := awssm.GetSecret(secretName)
	var syntaxError *json.SyntaxError
	assert.ErrorAs(err, &syntaxError, "error should be a syntax error")
}
