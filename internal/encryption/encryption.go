package encryption

import (
	"strings"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/compoundencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/gcmencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/noopencryptor"
)

type Encryptor interface {
	Encrypt(plaintext []byte) (string, error)
	Decrypt(ciphertext string) ([]byte, error)
}

func EncryptorFromKeys(keys ...string) Encryptor {
	switch len(keys) {
	case 0:
		return noopencryptor.NewNoopEncryptor()
	case 1:
		return encryptorFromKey(keys[0])
	default:
		var encryptors []compoundencryptor.Encryptor
		for _, key := range keys {
			encryptors = append(encryptors, encryptorFromKey(key))
		}
		return compoundencryptor.NewCompoundEncryptor(encryptors[0], encryptors[1:]...)
	}
}

func encryptorFromKey(key string) Encryptor {
	if (strings.TrimSpace(key) == key) && len(key) > 0 {
		var keyAs32ByteArray [32]byte
		copy(keyAs32ByteArray[:], key)
		return gcmencryptor.NewGCMEncryptor(&keyAs32ByteArray)
	}
	return noopencryptor.NewNoopEncryptor()
}
