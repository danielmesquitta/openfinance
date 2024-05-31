package crypto

type Encrypter interface {
	Decrypt(hashed string) (string, error)
	Encrypt(plaintext string) (string, error)
}
