package saltedpassword

import (
	"crypto/sha256"
	"errors"

	"golang.org/x/crypto/pbkdf2"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/gcmencryptor"
)

func New(label, password string, salt []byte) (SaltedPassword, error) {
	switch {
	case len(salt) != 32:
		return SaltedPassword{}, errors.New("salt must be 32 bytes")
	case len(password) < 20:
		return SaltedPassword{}, errors.New("password must be at least 20 characters")
	case len(password) > 1024:
		return SaltedPassword{}, errors.New("password must not be more than 1024 characters")
	}

	var key [32]byte
	copy(key[:], pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New))

	return SaltedPassword{
		Label:     label,
		Encryptor: gcmencryptor.NewGCMEncryptor(&key),
	}, nil
}

type SaltedPassword struct {
	Label     string
	Encryptor gcmencryptor.GCMEncryptor
}
