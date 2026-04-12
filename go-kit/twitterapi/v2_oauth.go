package twitterapi

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/duongtuttbn/shared-resource/go-kit/collections"
	"github.com/duongtuttbn/shared-resource/go-kit/kit"
	"github.com/duongtuttbn/shared-resource/go-kit/log"
)

var _ V2Oauth = (*v2OauthImpl)(nil)

type v2OauthImpl struct {
	V2
	httpClient *resty.Client
	cfg        Config
}

func NewV2OAuth(cfg Config, v2 ...V2) V2Oauth {
	var v2Service V2
	if len(v2) > 0 {
		v2Service = v2[0]
	} else {
		v2Service = NewV2()
	}
	return &v2OauthImpl{
		cfg: cfg,
		V2:  v2Service,
		httpClient: resty.New().
			SetRetryCount(5).
			SetRetryMaxWaitTime(15 * time.Minute).
			AddRetryCondition(func(response *resty.Response, err error) bool {
				if err != nil {
					return true
				}

				if response.StatusCode() == http.StatusTooManyRequests {
					log.Warnf("Response status code is %d - Body: %s - Retrying...", response.StatusCode(), response.Body())
					return true
				}

				return false
			}).
			SetRetryAfter(func(_ *resty.Client, response *resty.Response) (time.Duration, error) {
				if response.StatusCode() != http.StatusTooManyRequests {
					return defaultRetryWaitTime, nil
				}

				rateLimitReset := response.Header().Get("x-rate-limit-reset")
				if rateLimitReset == "" {
					return defaultRetryWaitTime, nil
				}

				rateLimitResetTime, err := strconv.ParseInt(rateLimitReset, 10, 64)
				if err != nil {
					return defaultRetryWaitTime, nil
				}

				seconds := rateLimitResetTime - time.Now().Unix() + 1
				retryAfter := time.Duration(seconds) * time.Second

				log.Warnf("Response status code is %d - Retrying after %s...", response.StatusCode(), retryAfter)

				return retryAfter, nil
			}).
			SetBaseURL(baseURL),
	}
}

func (t *v2OauthImpl) GetOAuthURL(state string, codeChallenge string, challengeMethod ChallengeMethod, additionalScopes ...Scope) string {
	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("client_id", string(t.cfg.ClientID))
	query.Set("redirect_uri", t.cfg.RedirectURI)
	query.Set("state", state)
	query.Set("code_challenge", codeChallenge)
	query.Set("code_challenge_method", string(challengeMethod))
	query.Set("scope", buildOauthScopeParam([]Scope{
		ScopeTweetRead,
		ScopeUsersRead,
		ScopeFollowsRead,
		ScopeOfflineAccess,
	}, additionalScopes))
	return "https://x.com/i/oauth2/authorize?" + query.Encode()
}

func (t *v2OauthImpl) GetAccessToken(ctx context.Context, code string, codeVerifier string) (*TokenResponse, error) {
	var accessTokenResult *TokenResponse
	resp, err := t.httpClient.R().
		SetContext(ctx).
		SetResult(&accessTokenResult).
		SetBasicAuth(string(t.cfg.ClientID), t.cfg.ClientSecret).
		SetHeaders(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		}).
		SetFormData(map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     string(t.cfg.ClientID),
			"redirect_uri":  t.cfg.RedirectURI,
			"code":          code,
			"code_verifier": codeVerifier,
		}).
		Post("/oauth2/token")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, errors.Errorf("Response code: %d - Body: %s", resp.StatusCode(), resp.String())
	}

	return accessTokenResult, nil
}

func (t *v2OauthImpl) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	var accessTokenResult *TokenResponse
	resp, err := t.httpClient.R().
		SetContext(ctx).
		SetResult(&accessTokenResult).
		SetBasicAuth(string(t.cfg.ClientID), t.cfg.ClientSecret).
		SetHeaders(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		}).
		SetFormData(map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": refreshToken,
		}).
		Post("/oauth2/token")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.Errorf("Response code: %d - Body: %s", resp.StatusCode(), resp.String())
	}

	return accessTokenResult, nil
}

const codeVerifierCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"

// GenerateCodeVerifierPair creates a random 43-128 character string for PKCE
func GenerateCodeVerifierPair(length int) (string, string, error) {
	if length < 43 || length > 128 {
		return "", "", fmt.Errorf("code verifier length must be between 43 and 128 characters")
	}

	verifier := make([]byte, length)
	for i := range verifier {
		verifier[i] = codeVerifierCharset[rand.N(len(codeVerifierCharset))]
	}

	h := sha256.Sum256(verifier)
	return string(verifier), base64.RawURLEncoding.EncodeToString(h[:]), nil
}

func buildOauthScopeParam(defaultScopes []Scope, additionalScopes []Scope) string {
	scopes := lo.Uniq(collections.Merge(defaultScopes, additionalScopes))
	return strings.Join(kit.Map(scopes, func(item Scope) string {
		return string(item)
	}), " ")
}
