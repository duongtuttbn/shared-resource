package hash

import (
	"crypto/md5"
)

func MD5(str []byte) []byte {
	h := md5.Sum(str)
	return h[:]
}
