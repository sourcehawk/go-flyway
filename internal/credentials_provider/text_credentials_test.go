package credentials_provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TextDatabaseCredentials_Validate_Succeeds(t *testing.T) {
	d := TextDatabaseCredentials{
		DatabaseCredentials: 
			DatabaseCredentials{
				Username: "a",
				Password: "b",
				Host:     "c",
				Port:     5432,
				Database: "d",
			},
	}
	assert := assert.New(t)
	assert.NoError(d.Validate())
}

func Test_TextDatabaseCredentials_Validate_FailsWithEmptyField(t *testing.T) {
	d := TextDatabaseCredentials{
		DatabaseCredentials: 
			DatabaseCredentials{
				Username: "a",
				Password: "b",
				Host:     "c",
				Port:     5432,
				Database: "d",
			},
	}
	assert := assert.New(t)
	d.Username = ""
	assert.Error(d.Validate())
	d.Username = "a"
	d.Password = ""
	assert.Error(d.Validate())
	d.Password = "b"
	d.Host = ""
	assert.Error(d.Validate())
	d.Host = "c"
	d.Port = 0
	assert.Error(d.Validate())
	d.Port = 5432
	d.Database = ""
	assert.Error(d.Validate())
	d.Database = "d"
	assert.NoError(d.Validate())
}

func Test_TextDatabaseCredentials_GetCredentials_Succeeds(t *testing.T) {
	d := TextDatabaseCredentials{
		DatabaseCredentials: 
			DatabaseCredentials{
				Username: "a",
				Password: "b",
				Host:     "c",
				Port:     5432,
				Database: "d",
			},
	}
	assert := assert.New(t)
	creds, err := d.GetCredentials()
	assert.NoError(err)
	assert.Equal(*creds, d.DatabaseCredentials)
}

func Test_TextDatabaseCredentials_GetCredentials_FailsOnValidationError(t *testing.T) {
	d := TextDatabaseCredentials{
		DatabaseCredentials: 
			DatabaseCredentials{
				Username: "a",
				Password: "b",
				Host:     "c",
				Port:     5432,
				// Database: "d",
			},
	}
	assert := assert.New(t)
	_, err := d.GetCredentials()
	assert.Error(err)
}
