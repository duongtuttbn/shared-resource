package jwtfx

import (
	"context"
	"errors"
	"time"
	"tla-backend/pkg/go-kit/core"
	"tla-backend/pkg/go-kit/dt"
	"tla-backend/pkg/go-kit/ids"
	"tla-backend/pkg/go-kit/kit"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/fx"
)

var (
	ErrExpired        = errors.New("token expired")
	ErrEncodeDisabled = errors.New("encode jwt is disabled")
	ErrDecodeDisabled = errors.New("decode jwt is disabled")
)

type JwtParams struct {
	fx.In
	Config  JwtConfig
	Decoder JwtDecoder `optional:"true"`
	Encoder JwtEncoder `optional:"true"`
}

type Claims struct {
	jwt.RegisteredClaims
	UserID dt.UserID              `json:"user_id"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

var _ core.AuthService = (*JwtService)(nil)

type JwtService struct {
	conf    JwtConfig
	encoder JwtEncoder
	decoder JwtDecoder
}

func NewJwtService(p JwtParams) *JwtService {
	return &JwtService{
		conf:    p.Config,
		encoder: p.Encoder,
		decoder: p.Decoder,
	}
}

func (s *JwtService) Verify(ctx context.Context, token string) (core.AuthInfo, error) {
	if s.decoder == nil {
		return nil, ErrDecodeDisabled
	}
	var claims Claims
	jwtToken, err := s.decoder.DecodeWithClaims(ctx, &claims, token)
	if err != nil {
		return nil, err
	}

	if !jwtToken.Valid ||
		(claims.ExpiresAt == nil || claims.ExpiresAt.Before(time.Now())) ||
		(claims.NotBefore != nil && claims.NotBefore.After(time.Now())) {
		return nil, ErrExpired
	}

	return &authInfo{
		userID: claims.UserID,
		data:   claims.Data,
	}, nil
}

func (s *JwtService) Sign(claims Claims, expireIn ...time.Duration) (string, error) {
	if s.encoder == nil {
		return "", ErrEncodeDisabled
	}

	duration := s.conf.JwtDefaultExpired
	if len(expireIn) > 0 {
		duration = expireIn[0]
	}
	now := time.Now()
	if claims.ID == "" {
		claims.ID = ids.NewUUID()
	}
	if claims.Issuer == "" {
		claims.Issuer = s.conf.JwtIssuer
	}
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.NotBefore = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(duration))
	mapClaims, err := kit.ConvertType[jwt.MapClaims](claims)
	if err != nil {
		return "", err
	}
	return s.encoder.Generate(mapClaims)
}

type authInfo struct {
	userID dt.UserID
	data   dt.Map
}

func (a *authInfo) Unauthorized() bool {
	return a == nil || a.userID == ""
}

func (a *authInfo) GetUserID() dt.UserID {
	if a == nil {
		return ""
	}
	return a.userID
}

func (a *authInfo) GetData() dt.Map {
	if a == nil {
		return nil
	}
	return a.data
}
