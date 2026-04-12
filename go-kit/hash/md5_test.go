package hash

import (
	"encoding/hex"
	"testing"
)

func TestMD5(t *testing.T) {
	str := "Hello, World!"
	expected := "65a8e27d8879283831b664bd8b7f0ad4"
	result := hex.EncodeToString(MD5([]byte(str)))
	if result != expected {
		t.Errorf("MD5(%s) = %s; expected %s", str, result, expected)
	}
}
