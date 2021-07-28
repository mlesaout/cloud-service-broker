package encryption

type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

var encryptorInstance Encryptor = nil

func SetEncryptor(encryptor Encryptor) {
	encryptorInstance = encryptor
}
