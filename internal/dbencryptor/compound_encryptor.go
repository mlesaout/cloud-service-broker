package dbencryptor

import "github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"

func NewCompoundEncryptor(primary models.Encryptor, secondaries ...models.Encryptor) models.Encryptor {
	return CompoundEncryptor{
		primary:     primary,
		secondaries: secondaries,
	}
}

type CompoundEncryptor struct {
	primary     models.Encryptor
	secondaries []models.Encryptor
}

func (c CompoundEncryptor) Encrypt(plaintext []byte) (string, error) {
	return c.primary.Encrypt(plaintext)
}

func (c CompoundEncryptor) Decrypt(ciphertext string) (data []byte, err error) {
	for _, decryptor := range append([]models.Encryptor{c.primary}, c.secondaries...) {
		data, err = decryptor.Decrypt(ciphertext)
		if err == nil {
			return data, nil
		}
	}

	return nil, err
}
