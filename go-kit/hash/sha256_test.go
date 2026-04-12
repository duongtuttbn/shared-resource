package hash

import (
	"encoding/hex"
	"testing"
)

func TestSHA256(t *testing.T) {
	str := "Hello, World!"
	expected := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	result := hex.EncodeToString(SHA256([]byte(str)))
	if result != expected {
		t.Errorf("SHA256(%s) = %s; expected %s", str, result, expected)
	}
}

func TestHmacSHA256(t *testing.T) {
	str := "Hello, World!"
	secret := "secret"
	expected := "fcfaffa7fef86515c7beb6b62d779fa4ccf092f2e61c164376054271252821ff"
	result := hex.EncodeToString(HmacSHA256([]byte(secret), []byte(str)))
	if result != expected {
		t.Errorf("SHA256(%s) = %s; expected %s", str, result, expected)
	}
}
