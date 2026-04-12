package aes

import (
	"encoding/base64"
	"slices"

	"github.com/pkg/errors"
)

type DataEncryptor interface {
	Encrypt(plain []byte) ([]byte, error)
	EncryptString(plain string) (string, error)
	Decrypt(encrypted []byte) ([]byte, error)
	DecryptString(encrypted string) (string, error)
}

type dataEncryptor struct {
	encryptionKey []byte
}

func NewDataEncryptor(encryptionKey []byte) DataEncryptor {
	return &dataEncryptor{
		encryptionKey: encryptionKey,
	}
}

func (s *dataEncryptor) Encrypt(plain []byte) ([]byte, error) {
	encrypted, nonce, err := Encrypt(plain, s.encryptionKey)
	if err != nil {
		return nil, err
	}

	return slices.Concat(nonce, encrypted), nil
}

func (s *dataEncryptor) EncryptString(plain string) (string, error) {
	encrypted, err := s.Encrypt([]byte(plain))
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(encrypted), nil
}

func (s *dataEncryptor) Decrypt(encrypted []byte) ([]byte, error) {
	if len(encrypted) < 12 {
		return nil, errors.New("invalid encrypted data")
	}
	nonce := encrypted[:12]
	encryptedData := encrypted[12:]
	plain, err := Decrypt(encryptedData, s.encryptionKey, nonce)
	if err != nil {
		return nil, err
	}

	return plain, nil
}

func (s *dataEncryptor) DecryptString(encrypted string) (string, error) {
	encryptedBytes, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	plain, err := s.Decrypt(encryptedBytes)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}
