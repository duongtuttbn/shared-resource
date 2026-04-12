package aes_test

import (
	"testing"
	"github.com/duongtuttbn/shared-resource/go-kit/aes"

	"github.com/stretchr/testify/require"
)

func TestEncryptor(t *testing.T) {
	data := []byte("test")
	key, err := aes.GenerateRandomBytes(32)
	require.NoError(t, err)
	cipher, nonce, err := aes.Encrypt(data, key)
	require.NoError(t, err)

	plaintext, err := aes.Decrypt(cipher, key, nonce)
	require.NoError(t, err)
	require.Equal(t, data, plaintext)
}
