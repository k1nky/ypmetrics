// Пакет crypto представляет инстуременты для асимметричного шифрования.
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
)

var (
	// ErrInvalidKeyFormat неверный формат ключа
	ErrInvalidKeyFormat = errors.New("invalid key format")
)

func ReadPrivateKey(r io.Reader) (*rsa.PrivateKey, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(content)
	if block == nil {
		return nil, ErrInvalidKeyFormat
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if _, ok := key.(*rsa.PrivateKey); !ok {
		return nil, ErrInvalidKeyFormat
	}
	return key.(*rsa.PrivateKey), err
}

func ReadPublicKey(r io.Reader) (*rsa.PublicKey, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(content)
	if block == nil {
		return nil, ErrInvalidKeyFormat
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if _, ok := key.(*rsa.PublicKey); !ok {
		return nil, ErrInvalidKeyFormat
	}
	return key.(*rsa.PublicKey), err
}

func DecryptRSA(key *rsa.PrivateKey, msg []byte) ([]byte, error) {
	var decrypted []byte

	chunkSize := key.Size()
	chunks := chunkBytes(msg, chunkSize)
	hash := sha256.New()
	rnd := rand.Reader

	for i := range chunks {
		b, err := rsa.DecryptOAEP(hash, rnd, key, chunks[i], nil)
		if err != nil {
			return nil, err
		}
		decrypted = append(decrypted, b...)
	}
	return decrypted, nil
}

func EncryptRSA(key *rsa.PublicKey, msg []byte) ([]byte, error) {
	var encrypted []byte

	chunkSize := key.Size() - sha256.Size*2 - 2
	chunks := chunkBytes(msg, chunkSize)
	hash := sha256.New()
	rnd := rand.Reader

	for i := range chunks {
		b, err := rsa.EncryptOAEP(hash, rnd, key, chunks[i], nil)
		if err != nil {
			return nil, err
		}
		encrypted = append(encrypted, b...)
	}
	return encrypted, nil
}

func chunkBytes(src []byte, chunkSize int) [][]byte {
	var chunks [][]byte

	for i := 0; i < len(src); i += chunkSize {
		end := i + chunkSize
		if end > len(src) {
			end = len(src)
		}
		chunks = append(chunks, src[i:end])
	}
	return chunks
}
