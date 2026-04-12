package jwtfx

import (
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
)

const jwkRsaPrivateKey = `{
    "p": "5Bm42im0PgiRvKhppZYHG_WAIf2XINp7dhp719La5oPjmdmvyA-TzTTX4EmVUt1xKtN8tfmrhbjgrsg3XtDfd4WT86Kbul3b6nCm9ioeiw28cqlNdw2bTdVIR2X1kr9mUpxAGRBy4hUSmYIBGBhhpbAKD9OSdg0yWQZmMuWUEzM",
    "kty": "RSA",
    "q": "lQY-dRpcAxsrmIA_VbCPqZITRE43Ky9IX3eChtE9Jrlk0x9ha20pdwl6cktfTyrhaoE9pvwvs6RpHp1pOMsoWKZMVyqwAFsPhIgEkJiryOLt68aKAho_m3dYw5eQsK8qyumA57hR2sOaLb1op48WUHhm5VogEsSwJLQJKtTuLAU",
    "d": "D1wXIEnNedf2Yo-lyynmchLDG7695WFiwu2h1L4cA7dpcVUOF43Hn6Zo1R51ejNKgZ-W5EuJm377KMvdhiE8DvNnlZPJAMmxMjfKB35a8TPac07mfYNzstwdVQuhrQZ5CwEO0Vk2fXZW2j_hn_wB2_2syWwxIjLNbi4LugRcPpt7quWuotuqbCdo6n6TszuE2zPiu7jxpO31Ben9oUpOytg6xw6InqIqB0342R3_2jyL22Jc5Kcx9Gy1q2Lq4ZBWODkL5-USYmaILRoNWK6i_141JbUTIIkn8U7-OOd8kO6Q_OkfoDicLhssO8gmzVIo2JLnTUvyxFx1U0KATIX2qQ",
    "e": "AQAB",
    "use": "sig",
    "kid": "4rYKldB778ftdh2WCsw2o5SfIF37V9pRN_ysnFG-Z5I",
    "qi": "Xg3cRJjD8DVdO2ApsYLIcxk-45rJCY8UG1-WW9UMI_LxUD0eDH0czPbPf1IubyLTdKec3h0j_zURZkYZcrxZtb9Wz366ESJPvaAXXSmhYSoCs5Ewovjx278hfVIbp4Ly7oMoX5auC1Xdgdma7tccVEEZ16V5dyTvQI0tr28QmKQ",
    "dp": "a3Taqpwe92JeFbxZGNLWwosjM-AdhDKpGvhbA0-oJBRZ8q6kquD7xh5w3I6NtB3yJDTBeZEHBtYTswNLYnWP8OSS0KH4LxHsekNbxHgPL37nGjU78ywLz9z8UfZsfBeDAsPtRmGDXZKD0qF2Fn3V8pI-CzqmssqAv4POPYf9_BU",
    "alg": "RS256",
    "dq": "GKtk8XvAmZ8I04D_ew70aUzONbOA_HwiTfN5vxmqNtvf7fc26FK014jRJVSG3ZMqp7fnXdpHh0SDRlcmkQlIj4xP_OoLIrPwWK8vmkQ7w9CVND-0nu57cyAJqK9Re34z5k1LUpC3tDBHOKUvSvWr6vxThEosHw9CXYEUN2vyVYU",
    "n": "hMiI1-FlzOxAyTHSGiBnmeyoHgyMwHT6tXxlQEJ7c8X1BDzmMAPBGoJuRb33_D5tpNPzB2jqdO1H0YBdifhxMt8Gr4IQBpMYeq5ScFPUFfn6OyQGOjcVdB-_vmLmHyPDZsLoEl8E_J4MPn8rHjXjs6rLUwCpDxeoxmgjqWuP2nIidCEZdPh9pZuq7YxAQjTYduKnBHEbpLc8qVyriqoAphn7JgYUj5Mt8C5Uk8S_fJn5LrZAslJyqC0AONYOl7cqlsNKDpqFYkQ7Dn4FfFCMIZpp6lP4_ZYbZNbJaH5gn0QE0efsMnBDGdOXBkogjWXeASBHv-PoBiqK9L9iBJsj_w"
}`

const jwkEdDSAPrivateKey = `{
    "kty": "OKP",
    "d": "hB9_C6VXPe2RhcotUxGMxpujiBawb2c5JL8ItIlWXsk",
    "use": "sig",
    "crv": "Ed25519",
    "kid": "47WxR6uaKt-mPtKtyg-SSEnCVW3AB9VV64dVTHGB5vQ",
    "x": "KsfoYiJ360dya5OcNzRJZ3xwyCCDieE7FlyBXfvh8cc",
    "alg": "EdDSA"
}`

func TestGenerateRSAJwt(t *testing.T) {
	jwk, err := jwkset.NewJWKFromRawJSON(json.RawMessage(jwkRsaPrivateKey), jwkset.JWKMarshalOptions{
		Private: true,
	}, jwkset.JWKValidateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	privateKey := jwk.Key().(*rsa.PrivateKey)
	encoder := newRsaEncoder(privateKey, jwk.Marshal().KID)
	token, err := encoder.Generate(jwt.MapClaims{"sub": "1234567890", "exp": time.Now().Add(time.Minute).Unix()})
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return privateKey.Public(), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if !parsed.Valid {
		t.Fatal("invalid token")
	}

	claims := parsed.Claims.(jwt.MapClaims)
	if claims["sub"] != "1234567890" {
		t.Fatal("invalid id")
	}
}

func TestGenerateEdDSAJwt(t *testing.T) {
	jwk, err := jwkset.NewJWKFromRawJSON(json.RawMessage(jwkEdDSAPrivateKey), jwkset.JWKMarshalOptions{
		Private: true,
	}, jwkset.JWKValidateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	privateKey := jwk.Key().(ed25519.PrivateKey)
	encoder := newEdDSAEncoder(privateKey, jwk.Marshal().KID)
	token, err := encoder.Generate(jwt.MapClaims{"sub": "1234567890", "exp": time.Now().Add(time.Minute).Unix()})
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return privateKey.Public(), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if !parsed.Valid {
		t.Fatal("invalid token")
	}

	claims := parsed.Claims.(jwt.MapClaims)
	if claims["sub"] != "1234567890" {
		t.Fatal("invalid id")
	}
}
