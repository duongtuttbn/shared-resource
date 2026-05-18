package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
)

type TelegramInfo struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

func VerifyTelegramAuthorization(data, token string) (*TelegramInfo, error) {
	params, _ := url.ParseQuery(data)
	var authData = make([]string, 0)
	var hash = ""
	for k, v := range params {
		if k == "hash" {
			hash = v[0]
			continue
		}
		authData = append(authData, k+"="+v[0])
	}
	sort.Strings(authData)
	var imploded = ""
	for _, s := range authData {
		if imploded != "" {
			imploded += "\n"
		}
		imploded += s
	}
	hashSecret := computeHmac256(token, "WebAppData")
	_hash := computeHmac256(imploded, string(hashSecret))

	if hash != hex.EncodeToString(_hash) {
		return nil, errors.New("unauthorized")
	}

	var userInfo *TelegramInfo
	err := json.Unmarshal([]byte(params["user"][0]), &userInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func computeHmac256(message string, secret string) []byte {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return h.Sum(nil)
}
