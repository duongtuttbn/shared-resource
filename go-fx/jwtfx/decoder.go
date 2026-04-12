package jwtfx

import (
	"context"
	"encoding/json"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

var (
	ErrInvalidTokens            = errors.New("invalid token")
	_                JwtDecoder = (*hmacDecoder)(nil)
	_                JwtDecoder = (*decoder)(nil)
	_                JwtDecoder = (*chainedDecoder)(nil)
)

type JwtDecoder interface {
	Decode(token string) (*jwt.Token, error)
	DecodeWithContext(ctx context.Context, token string) (*jwt.Token, error)
	DecodeWithClaims(ctx context.Context, claims jwt.Claims, token string) (*jwt.Token, error)
}

// NewJwtDecoder create a JwtDecoder.
func NewJwtDecoder(env JwtConfig) (JwtDecoder, error) {
	var decoders []JwtDecoder
	if env.JwtJwksJSON != "" {
		k, err := keyfunc.NewJWKSetJSON(json.RawMessage(env.JwtJwksJSON))
		if err != nil {
			return nil, err
		}
		d, err := newDecoder(k.KeyfuncCtx)
		if err != nil {
			return nil, err
		}
		decoders = append(decoders, d)
	}

	if env.JwtJwkPrivateKeyJSON != "" {
		k, err := keyfunc.NewJWKJSON(json.RawMessage(env.JwtJwkPrivateKeyJSON))
		if err != nil {
			return nil, err
		}
		d, err := newDecoder(k.KeyfuncCtx)
		if err != nil {
			return nil, err
		}
		decoders = append(decoders, d)
	}

	if env.JwtKey != "" {
		decoders = append(decoders, newHMACDecoder([]byte(env.JwtKey)))
	}

	if env.JwtJwksURL != "" {
		d, err := newDecoderWithJwksURL(env.JwtJwksURL)
		if err != nil {
			return nil, err
		}
		decoders = append(decoders, d)
	}

	if len(decoders) == 0 {
		return nil, errors.New("no jwt config provided for decoder")
	}

	return newChainedDecoder(decoders...), nil
}

type hmacDecoder struct {
	secret []byte
}

func newHMACDecoder(secret []byte) *hmacDecoder {
	return &hmacDecoder{
		secret: secret,
	}
}

func (d *hmacDecoder) Decode(token string) (*jwt.Token, error) {
	return d.DecodeWithContext(context.TODO(), token)
}

func (d *hmacDecoder) DecodeWithContext(ctx context.Context, token string) (*jwt.Token, error) {
	return d.DecodeWithClaims(ctx, jwt.MapClaims{}, token)
}

func (d *hmacDecoder) DecodeWithClaims(_ context.Context, claims jwt.Claims, token string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidTokens
		}
		return d.secret, nil
	})
}

type decoder struct {
	keyFuncCtx func(context.Context) jwt.Keyfunc
}

func newDecoder(keyFuncCtx func(context.Context) jwt.Keyfunc) (*decoder, error) {
	return &decoder{
		keyFuncCtx: keyFuncCtx,
	}, nil
}

func newDecoderWithJwksURL(jwksURL string) (*decoder, error) {
	k, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		return nil, err
	}
	return &decoder{
		keyFuncCtx: k.KeyfuncCtx,
	}, nil
}

func (d *decoder) Decode(token string) (*jwt.Token, error) {
	return d.DecodeWithContext(context.TODO(), token)
}

func (d *decoder) DecodeWithContext(ctx context.Context, token string) (*jwt.Token, error) {
	return d.DecodeWithClaims(ctx, jwt.MapClaims{}, token)
}

func (d *decoder) DecodeWithClaims(ctx context.Context, claims jwt.Claims, token string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		v, err := d.keyFuncCtx(ctx)(token)
		if err != nil {
			if errors.Is(err, keyfunc.ErrKeyfunc) {
				return nil, ErrInvalidTokens
			}
			return nil, err
		}

		return v, nil
	})
}

type chainedDecoder struct {
	decoders []JwtDecoder
}

func newChainedDecoder(decoders ...JwtDecoder) *chainedDecoder {
	return &chainedDecoder{
		decoders: decoders,
	}
}

func (d *chainedDecoder) Decode(token string) (*jwt.Token, error) {
	return d.DecodeWithContext(context.TODO(), token)
}

func (d *chainedDecoder) DecodeWithContext(ctx context.Context, token string) (*jwt.Token, error) {
	return d.DecodeWithClaims(ctx, jwt.MapClaims{}, token)
}

func (d *chainedDecoder) DecodeWithClaims(ctx context.Context, claims jwt.Claims, token string) (*jwt.Token, error) {
	for i := 0; i < len(d.decoders); i++ {
		t, err := d.decoders[i].DecodeWithClaims(ctx, claims, token)
		if err != nil {
			if errors.Is(err, ErrInvalidTokens) {
				continue
			}
			return nil, err
		}
		return t, nil
	}

	return nil, ErrInvalidTokens
}
