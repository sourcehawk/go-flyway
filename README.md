[![codecov](https://codecov.io/gh/sourcehawk/go-flyway/graph/badge.svg?token=14L5SEPDBW)](https://codecov.io/gh/sourcehawk/go-flyway)

# Go Flyway

go-flyway is a go wrapper for flyway migrations with secret store provider support for database credentials.

Currently it supports AWS Secrets Manager, environment variables and plain text credentials but a new provider can easily be plugged in with a small implementation. Feel free to open a PR or issue if you need a new provider.

## Installation

The migrator is provided as a go binary or docker image.

### Go

Note that the binary does not include the flyway CLI. You need to install it separately to run the binary. The migrator will use the flyway CLI installed in your PATH.

```bash
go install github.com/sourcehawk/go-flyway@latest
```

### Docker

The current docker image has version 11.8.2 of flyway CLI installed. You can use this image as a base image for your own docker image or run it directly. If you need a specific flyway version, [install the go-flyway binary](#go) and the flyway CLI separately into your own image.

```bash
docker pull ghcr.io/sourcehawk/go-flyway:latest
```

## Usage

The migrator takes a configuration file as input.

```bash
go-flyway --config ./path/to/config.yaml
```

Multiple configuration files can be provided. However, the schemas will not be merged but overwritten if the `schemas` field is defined in multiple files. The merging gives precedence to the last file in the list.

```bash
go-flyway --config ./config.yaml --config ./overwrite.yaml
```

If you are using the docker image, here is a docker compose example to run the migrator:

```yaml
services:
  migrator-db:
    image: postgres:17
    container_name: migrator-db
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: postgres

  migrator:
    image: ghcr.io/sourcehawk/go-flyway:latest
    container_name: go-flyway
    command:
      - "--config=/config.yml"
    volumes:
      # Ensure that the paths defined in the migrator config file are
      # correctly defined according to these mounts, e.g migrationsPath: /migrations/...
      - ./local/path/to/migrations:/migrations
      - ./local/path/to/config.yml:/config.yml
    depends_on:
      - migrator-db
```

## Configuration

The migrator is configured with a yaml file which has the following structure.

```yaml
# Default flyway arguments for all schemas (optional)
# If a schema defines the same argument with a different value,
# the schema's value will be used.
# Note: some arguments are managed by the migrator itself and should not be
# defined here. This includes arguments such as the schema name, migrations path,
# placeholders and credentials.
flywayArgs:
  - -connectRetries=10
  - -baselineOnMigrate=true
  - -outOfOrder=false
  - -validateMigrationNaming=true

# Default connection credentials for all schemas (optional)
# If the schema defines a `credentials` section, the schema's credentials will be used
credentials:
  # The provider to use for retrieving the credentials
  # Available providers are: "aws_sm", "text", "env"
  provider: aws_sm
  # The following configuration key should match the provider name (e.g. aws_sm in this case)
  aws_sm:
    # How to retrieve each required credentials field
    # The secretName is the name of the secret in AWS Secrets Manager
    # The secretKey is the key within that secret which contains the value for the field
    username:
      secretName: name/of/secret
      secretKey: username
    password:
      secretName: name/of/secret
      secretKey: password
    host:
      secretName: name/of/secret
      secretKey: host
    port:
      secretName: name/of/secret
      secretKey: port
    database:
      secretName: plant-hub/ci-smoke-test
      secretKey: database

# The schemas to be migrated will be processed in the order they are defined here.
# schemas[0] will be migrated first, then schemas[1], and so on
schemas:
  # The name of the schema to be migrated
  - name: schema_name
    # The path to the migrations directory for this schema
    migrationsPath: ./path/to/migrations
    # Placeholder values to be used for this schema (optional)
    # More information on placeholders can be found in the flyway documentation
    # https://www.red-gate.com/hub/product-learning/flyway/passing-parameters-and-settings-to-flyway-scripts
    placeholders:
      # The name of the placeholder as used in the migration scripts
      # E.g this one would be used as ${my_placeholder} in the migration scripts
      # Note that either the value or valueFromFile must be defined
      # If both are defined (which they should not), the valueFromFile will be used
      - name: my_placeholder
        # The value to be used for this placeholder (optional)
        # This value can be used in the migration scripts as ${my_placeholder}
        value: my_value
      # A placeholder that gets its value from a file
      - name: my_placeholder_from_file
        # The value to be used for this placeholder (optional)
        # This value can be used in the migration scripts as ${my_placeholder_from_file}
        # The value must be a path to a file that contains the value
        # The file will be read and the contents will be used as the value
        valueFromFile: ./path/to/file
    # Flyway arguments for this schema (optional)
    # If the argument, e.g 'connectRetries' is also defined in the top level
    # flywayArgs section, the schema's value will take precedence
    flywayArgs:
      - -connectRetries=10
      - -baselineOnMigrate=true
      - -outOfOrder=false
      - -validateMigrationNaming=true
    # The credentials to be used for this schema (optional)
    # If the credentials are defined in the top level credentials section,
    # the schema's credentials will take precedence
    credentials:
      # The provider to use for retrieving the credentials
      # Available providers are: "aws_sm", "text", "env"
      provider: text
      # The following configuration key should match the provider name (e.g. text in this case)
      text:
        username: <username>
        password: <password>
        host: <host>
        port: <port>
        database: <database>
```

### Credentials

The credentials section defines the credentials to be used for the migration. The credentials can be retrieved from different providers. The credentials can be defined both in the top level of the configuration file or in the schema section. If the credentials are defined in the schema section, they will override the top level credentials.

#### AWS Secrets Manager Credentials

Retrives the credentials from AWS Secrets Manager. The secrets must be in the format of a JSON objects.

```yaml
credentials:
  # The provider to use for retrieving the credentials
  # Available providers are: "aws_sm", "text", "env"
  provider: aws_sm
  # The following configuration key should match the provider name (e.g. aws_sm in this case)
  aws_sm:
    # How to retrieve each required credentials field
    # The secretName is the name of the secret in AWS Secrets Manager
    # The secretKey is the key within that secret which contains the value for the field
    username:
      secretName: plant-hub/ci-smoke-test
      secretKey: username
    password:
      secretName: plant-hub/ci-smoke-test
      secretKey: password
    host:
      secretName: plant-hub/ci-smoke-test
      secretKey: host
    port:
      secretName: plant-hub/ci-smoke-test
      secretKey: port
    database:
      secretName: plant-hub/ci-smoke-test
      secretKey: database
```

#### Environment Variables Credentials

Retrieves the credentials from environment variables, safer than plain text credentials but not as safe as secret stores such as AWS Secrets Manager. The environment variables must be set in the environment where the migrator is running.

```yaml
credentials:
  # The provider to use for retrieving the credentials
  # Available providers are: "aws_sm", "text", "env"
  provider: env
  # The following configuration key should match the provider name (e.g. env in this case)
  env:
    # Each of the following keys is the name of the environment variable
    # The value of the environment variable will be used as the value for the field
    # The environment variable must be set in the environment where the migrator is running
    usernameKey: MY_USERNAME_ENV_VAR
    passwordKey: MY_PASSWORD_ENV_VAR
    hostKey: MY_HOST_ENV_VAR
    portKey: MY_PORT_ENV_VAR
    databaseKey: MY_DATABASE_ENV_VAR
```

#### Plain Text Credentials

Plain text credentials are used for testing purposes only. The credentials are defined in the configuration file and are not retrieved from any provider. This is useful for local testing or for testing in environments where the credentials are not stored in a secrets manager.

```yaml
credentials:
  # The provider to use for retrieving the credentials
  # Available providers are: "aws_sm", "text", "env"
  provider: text
  # The following configuration key should match the provider name (e.g. text in this case)
  text:
    username: <username>
    password: <password>
    host: <host>
    port: <port>
    database: <database>
```

## Development

### Local testing

Testing the migrator

```bash
go test -v ./...
```

The migrator is also validated using [golangci-lint](https://golangci-lint.run/welcome/quick-start/). To run the linter locally, run the following command:

```bash
golangci-lint run
```

To test migrator against flyway migrations (integration tests), run the following commands and watch for any non-zero exit codes:

```bash
docker compose down
docker compose up --build
```

## TBD

- Set up AWS OIDC provider and role for GitHub Actions to test against AWS Secrets Manager
- Add support for other secret stores (e.g. HashiCorp Vault, Azure Key Vault, GCP Secret Manager)
