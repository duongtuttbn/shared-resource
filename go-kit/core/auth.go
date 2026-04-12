package core

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"tla-backend/pkg/go-kit/dt"
	"tla-backend/pkg/go-kit/kit"
	"tla-backend/pkg/go-kit/lerror"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	authInfoContextKey  = "auth_info"
	apiKeyUserIDPrefix  = "apiKey:"
)

type AuthInfo interface {
	Unauthorized() bool
	GetUserID() dt.UserID
	GetData() dt.Map
}

var _ AuthInfo = (*authInfo)(nil)

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

type AuthService interface {
	Verify(ctx context.Context, token string) (AuthInfo, error)
}

var _ AuthService = (*apiKeyAuth)(nil)

type apiKeyAuth struct {
	apiKeys []string
}

func NewAPIKeyAuth(apiKeys []string) AuthService {
	return &apiKeyAuth{apiKeys: apiKeys}
}

func (a *apiKeyAuth) Verify(_ context.Context, token string) (AuthInfo, error) {
	if !slices.Contains(a.apiKeys, token) {
		return nil, errors.New("invalid api key")
	}

	return &authInfo{
		userID: dt.UserID(fmt.Sprintf("%s%s", apiKeyUserIDPrefix, token)),
	}, nil
}

func UseAuth(authService AuthService, overridePrefix ...string) gin.HandlerFunc {
	return UseAuthWithHeader(authService, authorizationHeader, overridePrefix...)
}

func UseAuthWithHeader(authService AuthService, authHeaderName string, overridePrefix ...string) gin.HandlerFunc {
	prefix := bearerPrefix
	if len(overridePrefix) > 0 {
		prefix = overridePrefix[0]
	}

	return func(c *gin.Context) {
		_, existed := c.Get(authInfoContextKey)
		if existed {
			c.Next()
			return
		}

		reqToken := c.GetHeader(authHeaderName)
		if len(reqToken) == 0 {
			c.Next()
			return
		}

		if len(prefix) > 0 {
			if !strings.HasPrefix(reqToken, prefix) {
				c.Next()
				return
			}
			reqToken = strings.TrimPrefix(reqToken, prefix)
		}

		info, err := authService.Verify(c, reqToken)
		if err != nil || info == nil {
			c.Next()
			return
		}

		c.Set(authInfoContextKey, info)
		c.Next()
	}
}

func GetAuthInfo(c context.Context) AuthInfo {
	userInfo, existed := c.Value(authInfoContextKey).(AuthInfo)
	if !existed {
		return (*authInfo)(nil)
	}
	return userInfo
}

func GetAuthData(ctx context.Context) dt.Map {
	userInfo, ok := ctx.Value(authInfoContextKey).(AuthInfo)
	if !ok {
		return nil
	}

	return userInfo.GetData()
}

func GetAuthDataT[T any](ctx context.Context) T {
	data := GetAuthData(ctx)
	if data == nil {
		var zero T
		return zero
	}

	v, err := kit.ConvertType[T](data)
	if err != nil {
		var zero T
		return zero
	}

	return v
}

func FillAuthInfo(c context.Context, authParam *dt.AuthParam) error {
	userInfo := GetAuthInfo(c)
	if userInfo.Unauthorized() {
		return lerror.Unauthorized.ToError()
	}

	authParam.AuthUserID = userInfo.GetUserID()
	return nil
}

func RequireAuth(c *gin.Context) {
	userInfo := GetAuthInfo(c)
	if userInfo.Unauthorized() {
		c.Abort()
		ErrorHandler(c, lerror.Unauthorized.ToError())
		return
	}
}

func OnlyAPIKeyAuth(c *gin.Context) {
	userInfo := GetAuthInfo(c)
	if !strings.HasPrefix(string(userInfo.GetUserID()), apiKeyUserIDPrefix) {
		c.Abort()
		ErrorHandler(c, lerror.PermissionDenied.ToError())
		return
	}
}

func DisableAPIKeyAuth(c *gin.Context) {
	userInfo := GetAuthInfo(c)
	if strings.HasPrefix(string(userInfo.GetUserID()), apiKeyUserIDPrefix) {
		c.Abort()
		ErrorHandler(c, lerror.PermissionDenied.ToError())
		return
	}
}
