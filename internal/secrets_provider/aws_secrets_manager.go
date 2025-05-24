package secrets_provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go"
)

var RequestTimeoutDuration time.Duration = 10 * time.Second

type SecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type AWSSecretsManager struct {
	client SecretsManagerClient
}

func NewAWSSecretsManager() (*AWSSecretsManager, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeoutDuration)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return nil, fmt.Errorf("unable to load aws auth configuration: %w", err)
	}

	return &AWSSecretsManager{
		client: secretsmanager.NewFromConfig(cfg),
	}, nil
}

func (s *AWSSecretsManager) GetSecret(name string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeoutDuration)
	defer cancel()

	resp, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})

	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			return nil, fmt.Errorf("failed to get secret %s: %w", name, err)
		}
		return nil, fmt.Errorf("failed to get secret %s: %w", name, err)
	}

	var data map[string]any

	if resp.SecretString != nil {
		if err := json.Unmarshal([]byte(*resp.SecretString), &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal json from secret string: %w", err)
		}
	} else {
		if err := json.Unmarshal(resp.SecretBinary, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal json from secret binary: %w", err)
		}
	}

	return data, nil
}
