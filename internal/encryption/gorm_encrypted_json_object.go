package encryption

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSON interface{}

type EncryptedJSON map[string]interface{}

func (e *EncryptedJSON) Scan(value interface{}) error {
	encrypted, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("could not cast database value to byte slice")
	}

	decrypted, err := encryptorInstance.Decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("could not decrypt")
	}

	if err := json.Unmarshal(decrypted, e); err != nil {
		return fmt.Errorf("failed to unmarshal")
	}

	return nil
}

// Value return json value, implement driver.Valuer interface
func (e EncryptedJSON) Value() (driver.Value, error) {
	marshalled, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall")
	}

	encrypted, err := encryptorInstance.Encrypt(marshalled)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt")
	}

	return encrypted, nil
}
