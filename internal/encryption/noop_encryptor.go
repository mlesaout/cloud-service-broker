package encryption

type NoopEncryptor struct{}

func NewNoopEncryptor() NoopEncryptor {
	return NoopEncryptor{}
}

func (d NoopEncryptor) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	return plaintext, nil
}

func (d NoopEncryptor) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	return ciphertext, nil
}
