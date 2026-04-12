package jwtfx

import (
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"
	"errors"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
)

// JwtEncoder generate token that expire after time.Duration.
type JwtEncoder interface {
	Generate(claims jwt.MapClaims) (string, error)
}

// NewJwtEncoder create a JwtEncoder.
func NewJwtEncoder(env JwtConfig) (JwtEncoder, error) {
	switch {
	case env.JwtJwkPrivateKeyJSON != "":
		jwk, err := jwkset.NewJWKFromRawJSON(
			json.RawMessage(env.JwtJwkPrivateKeyJSON),
			jwkset.JWKMarshalOptions{Private: true},
			jwkset.JWKValidateOptions{},
		)
		if err != nil {
			return nil, err
		}
		switch privateKey := jwk.Key().(type) {
		case *rsa.PrivateKey:
			return newRsaEncoder(privateKey, jwk.Marshal().KID), nil
		case ed25519.PrivateKey:
			return newEdDSAEncoder(privateKey, jwk.Marshal().KID), nil
		default:
			return nil, errors.New("jwk is not an rsa or ed25519 private key")
		}
	case env.JwtKey != "":
		return newHmacEncoder([]byte(env.JwtKey)), nil
	default:
		return nil, errors.New("no jwt config provided for encoder")
	}
}

// hmacEncoder is a JwtEncoder that sign using HS256.
type hmacEncoder struct {
	secret []byte
}

func newHmacEncoder(secret []byte) *hmacEncoder {
	return &hmacEncoder{
		secret: secret,
	}
}

func (e *hmacEncoder) Generate(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(e.secret)
}

type rsaEncoder struct {
	privateKey *rsa.PrivateKey
	kid        string
}

func newRsaEncoder(privateKey *rsa.PrivateKey, kid string) *rsaEncoder {
	return &rsaEncoder{
		privateKey: privateKey,
		kid:        kid,
	}
}

func (e *rsaEncoder) Generate(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header[jwkset.HeaderKID] = e.kid
	return token.SignedString(e.privateKey)
}

type edDSAEncoder struct {
	privateKey ed25519.PrivateKey
	kid        string
}

func newEdDSAEncoder(privateKey ed25519.PrivateKey, kid string) *edDSAEncoder {
	return &edDSAEncoder{
		privateKey: privateKey,
		kid:        kid,
	}
}

func (e *edDSAEncoder) Generate(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header[jwkset.HeaderKID] = e.kid
	return token.SignedString(e.privateKey)
}
