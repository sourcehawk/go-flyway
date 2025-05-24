package secrets_provider

import "fmt"

type SecretsProvider interface {
	GetSecret(name string) (map[string]any, error)
}

type SecretRef struct {
	SecretName string `yaml:"secretName"`
	SecretKey  string `yaml:"secretKey,omitempty"`
}

func (s *SecretRef) Validate() error {
	if s.SecretName == "" {
		return fmt.Errorf("secretRef missing 'secretName' attribute")
	}
	if s.SecretKey == "" {
		return fmt.Errorf("secretRef '%s' missing 'secretKey' attribute", s.SecretName)
	}
	return nil
}

// SecretRefToStructJsonField links a JSON field name in a generic map
// to a specific key in a named secret, fetching the secret if needed.
type SecretRefToStructJsonField struct {
	// The JSON representation of a struct field name
	StructJsonField string
	// The secret reference that refers to this field of the struct
	SecretRef *SecretRef
}

// PopulateJSONFieldFromSecret fetches (once) the secret identified by
// SecretRef.SecretName from the provider `sp` (caching it in `secrets`),
// then looks up SecretRef.SecretKey inside that secret, and finally
// writes the resulting value into `jsonStruct` under StructJsonField.
//
// Panics if either `secrets` or `jsonStruct` is nil. Returns an error if
// the secret fetch fails or the specified key is missing/nil.
//
// After populating all the JSON fields your struct requires, you can then
// marshal the jsonStruct map to a byte array and then unmarshal that back
// into your concrete struct type, using the json library to do type conversions.
//
//	sp:         the secrets provider implementation
//	secrets:    cache of fetched secrets, keyed by secret name
//	jsonStruct: target map into which the secret value is inserted
func (s *SecretRefToStructJsonField) PopulateJSONFieldFromSecret(sp SecretsProvider, secrets map[string]map[string]any, jsonStruct map[string]any) error {
	if secrets == nil {
		panic("secrets map must be initialized")
	}

	if jsonStruct == nil {
		panic("jsonStruct must be initialized")
	}

	if s.SecretRef == nil {
		return fmt.Errorf("secret ref for field '%s' is missing (nil)", s.StructJsonField)
	}

	var secret map[string]any

	secret, ok := secrets[s.SecretRef.SecretName]

	if !ok {
		newSecret, err := sp.GetSecret(s.SecretRef.SecretName)

		if err != nil {
			return err
		}
		// update the secrets map with the newly fetched secret
		secrets[s.SecretRef.SecretName] = newSecret
		secret = newSecret
	}

	vAny, ok := secret[s.SecretRef.SecretKey]

	if !ok {
		return fmt.Errorf("key '%s' not present in secret %s", s.SecretRef.SecretKey, s.SecretRef.SecretName)
	}

	if vAny == nil {
		return fmt.Errorf("key '%s' in secret %s is nil", s.SecretRef.SecretKey, s.SecretRef.SecretName)
	}

	jsonStruct[s.StructJsonField] = vAny

	return nil
}
