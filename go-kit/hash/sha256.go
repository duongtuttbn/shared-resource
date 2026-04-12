package hash

import (
	"crypto/hmac"
	"crypto/sha256"
)

func SHA256(input []byte) []byte {
	h := sha256.Sum256(input)
	return h[:]
}

func HmacSHA256(secret []byte, input []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(input)

	return h.Sum(nil)
}
