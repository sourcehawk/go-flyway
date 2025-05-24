package migrator

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func Test_Placeholder_Validate_ReturnsWithoutErrorWithValueSet(t *testing.T) {
	p := Placeholder{
		Name:  "test",
		Value: "1",
	}

	assert := assert.New(t)
	assert.NoError(p.Validate())
}

func Test_Placeholder_Validate_ReturnsWithoutErrorWithValueFromFileSet(t *testing.T) {
	p := Placeholder{
		Name:          "test",
		ValueFromFile: "1",
	}

	assert := assert.New(t)
	assert.NoError(p.Validate())
}

func Test_Placeholder_Validate_FailsWithMissingName(t *testing.T) {
	p := Placeholder{
		ValueFromFile: "1",
	}

	assert := assert.New(t)
	assert.Error(p.Validate())
}

func Test_Placeholder_Validate_FailsWithMissingValueAndValueFromFile(t *testing.T) {
	p := Placeholder{
		Name: "test",
	}

	assert := assert.New(t)
	assert.Error(p.Validate())
}

func Test_Placeholder_Validate_FromYamlSucceeds(t *testing.T) {
	assert := assert.New(t)

	data := `
name: test
value: 10
`
	p := &Placeholder{}

	assert.NoError(yaml.Unmarshal([]byte(data), p))
	assert.Equal(p.Name, "test")
	assert.Equal(p.Value, "10")
}

func Test_Placeholder_Validate_FromYamlSucceedsWhenValueFromFileSpecified(t *testing.T) {
	assert := assert.New(t)

	data := `
name: test
valueFromFile: /test/file
`
	p := &Placeholder{}
	assert.NoError(yaml.Unmarshal([]byte(data), p))
	assert.Equal(p.Name, "test")
	assert.Equal(p.ValueFromFile, "/test/file")
}

func writeTestFile(name string, data []byte) (string, error) {
	tmpfile, err := os.CreateTemp("", name)

	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}

	if data != nil {
		_, err = tmpfile.Write(data)

		if err != nil {
			return "", fmt.Errorf("Error writing to temporary file: %v", err)
		}
	}

	err = tmpfile.Close()
	if err != nil {
		return "", fmt.Errorf("Error closing temporary file: %v", err)
	}

	return tmpfile.Name(), nil
}

func Test_Placeholder_loadValueFromFile_Succeeds(t *testing.T) {
	path, err := writeTestFile("testfile", []byte("test-data"))
	defer os.Remove(path) //nolint:errcheck

	assert := assert.New(t)
	assert.NoError(err)

	p := &Placeholder{
		Name:          "test",
		ValueFromFile: path,
	}

	assert.NoError(p.loadValueFromFile())
	assert.Equal(p.Value, "test-data")
}

func Test_Placeholder_loadValueFromFile_FailsIfValueFromFileNotSet(t *testing.T) {
	path, err := writeTestFile("testfile", nil)
	defer os.Remove(path) //nolint:errcheck

	assert := assert.New(t)
	assert.NoError(err)

	p := &Placeholder{
		Name: "test",
	}

	assert.Panics(func() { p.loadValueFromFile() }) //nolint:errcheck
}

func Test_Placeholder_loadValueFromFile_FailsIfFileEmpty(t *testing.T) {
	path, err := writeTestFile("testfile", nil)
	defer os.Remove(path) //nolint:errcheck

	assert := assert.New(t)
	assert.NoError(err)

	p := &Placeholder{
		Name:          "test",
		ValueFromFile: path,
	}

	assert.Error(p.loadValueFromFile())
}

func Test_Placeholder_loadValueFromFile_FailsIfFileDoesNotExist(t *testing.T) {

	p := &Placeholder{
		Name:          "test",
		ValueFromFile: "doesnotexistpath",
	}

	assert := assert.New(t)
	assert.Error(p.loadValueFromFile())
}

func Test_Placeholder_ToFlywayEnv_GetsArgRepresentationFromValue(t *testing.T) {
	p := &Placeholder{
		Name:  "test",
		Value: "test-123",
	}

	assert := assert.New(t)
	str, err := p.ToFlywayArg()
	assert.NoError(err)
	assert.Equal(str, "-placeholders.test=test-123")
}

func Test_Placeholder_ToFlywayEnv_GetsArgRepresentationFromFile(t *testing.T) {
	path, err := writeTestFile("testfile", []byte("test-data"))
	defer os.Remove(path) //nolint:errcheck

	assert := assert.New(t)
	assert.NoError(err)

	p := &Placeholder{
		Name:          "test",
		ValueFromFile: path,
	}
	v, err := p.ToFlywayArg()
	assert.NoError(err)
	assert.Equal(v, fmt.Sprintf("-placeholders.test=%s", "test-data"))
}

func Test_Placeholder_ToFlywayEnv_ReturnsVariableFromFileMultiline(t *testing.T) {
	data := `(
	'pot1', 'pot2', 'pot3',
	'pot4, 'pot5, 'pot6'
)`
	path, err := writeTestFile("testfile", []byte(data))
	defer os.Remove(path) //nolint:errcheck

	assert := assert.New(t)
	assert.NoError(err)

	p := &Placeholder{
		Name:          "test",
		ValueFromFile: path,
	}

	v, err := p.ToFlywayArg()
	assert.NoError(err)
	assert.Equal(v, fmt.Sprintf("-placeholders.test=%s", data))
}

func Test_Placeholder_ToFlywayEnv_FailsOnValidationError(t *testing.T) {
	p := &Placeholder{
		Name: "test",
	}

	assert := assert.New(t)
	_, err := p.ToFlywayArg()
	assert.Error(err)
}

func Test_Placeholder_ToFlywayEnv_FailsOnFileLoadError(t *testing.T) {
	p := &Placeholder{
		Name:          "test",
		ValueFromFile: "doesnotexistfile",
	}

	assert := assert.New(t)
	_, err := p.ToFlywayArg()
	assert.Error(err)
}
