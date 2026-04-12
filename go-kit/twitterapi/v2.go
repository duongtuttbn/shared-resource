package twitterapi

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tla-backend/pkg/go-kit/kit"
	"tla-backend/pkg/go-kit/log"
)

var _ V2 = (*v2Impl)(nil)

type v2Impl struct {
	httpClient *resty.Client
}

func NewV2(maxWaitTimeForRetry ...time.Duration) V2 {
	return &v2Impl{
		httpClient: resty.New().
			SetRetryCount(1).
			SetRetryMaxWaitTime(15 * time.Minute).
			AddRetryCondition(func(response *resty.Response, err error) bool {
				if err != nil {
					return true
				}

				if response.StatusCode() == http.StatusTooManyRequests {
					retryAfter, _ := getRetryAfter(response)
					if len(maxWaitTimeForRetry) > 0 && retryAfter > maxWaitTimeForRetry[0] {
						return false
					}
					log.Warnf("Response status code is %d - Body: %s - Retrying...", response.StatusCode(), response.Body())
					return true
				}

				return false
			}).
			SetRetryAfter(func(_ *resty.Client, response *resty.Response) (time.Duration, error) {
				if response.StatusCode() != http.StatusTooManyRequests {
					return defaultRetryWaitTime, nil
				}

				retryAfter, err := getRetryAfter(response)
				if err != nil {
					return defaultRetryWaitTime, err
				}

				log.Warnf("Response status code is %d - Retrying after %s...", response.StatusCode(), retryAfter)

				return retryAfter, nil
			}).
			SetBaseURL(baseURL),
	}
}

func getRetryAfter(response *resty.Response) (time.Duration, error) {
	rateLimitReset := response.Header().Get("x-rate-limit-reset")
	if rateLimitReset == "" {
		return defaultRetryWaitTime, nil
	}

	rateLimitResetTime, err := strconv.ParseInt(rateLimitReset, 10, 64)
	if err != nil {
		return defaultRetryWaitTime, nil
	}

	seconds := rateLimitResetTime - time.Now().Unix() + 1
	return time.Duration(seconds) * time.Second, nil
}

func (t *v2Impl) GetUserInfo(ctx context.Context, accessToken AccessToken) (*User, error) {
	return t.GetUserByID(ctx, accessToken, "me")
}

func (t *v2Impl) GetUserByID(ctx context.Context, accessToken AccessToken, id string) (*User, error) {
	var result APIV2Response[*User]
	resp, err := t.httpClient.
		R().
		SetContext(ctx).
		SetAuthToken(string(accessToken)).
		SetPathParam("id", id).
		SetQueryParam("user.fields", UserFields).
		SetResult(&result).
		Get("/users/{id}")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.Errorf("Response code: %d - Body: %s", resp.StatusCode(), resp.String())
	}

	if result.Data == nil {
		return nil, errors.New("user not found")
	}

	user := transformUser(*result.Data)
	return &user, nil
}

func (t *v2Impl) GetUserByIDs(ctx context.Context, accessToken AccessToken, ids []string) ([]User, error) {
	var result APIV2Response[[]User]
	resp, err := t.httpClient.
		R().
		SetContext(ctx).
		SetAuthToken(string(accessToken)).
		SetQueryParam("ids", strings.Join(ids, ",")).
		SetQueryParam("user.fields", UserFields).
		SetResult(&result).
		Get("/users")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.Errorf("Response code: %d - Body: %s", resp.StatusCode(), resp.String())
	}

	return kit.Map(result.Data, transformUser), nil
}

func (t *v2Impl) CreateTweet(ctx context.Context, accessToken AccessToken, param CreateTweetParam) (*SimpleTweet, error) {
	var response APIV2Response[*SimpleTweet]
	resp, err := t.httpClient.R().
		SetContext(ctx).
		SetResult(&response).
		SetAuthToken(string(accessToken)).
		SetBody(param).
		Post("/tweets")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if resp.IsError() {
		return nil, errors.Errorf("request failed with status: %d - Body: %s", resp.StatusCode(), resp.Body())
	}

	return response.Data, nil
}

func (t *v2Impl) GetTweet(ctx context.Context, accessToken AccessToken, id string) (*Tweet, error) {
	var response APIV2Response[Tweet]
	resp, err := t.httpClient.R().
		SetContext(ctx).
		SetResult(&response).
		SetAuthToken(string(accessToken)).
		SetPathParam("id", id).
		SetQueryParam("tweet.fields", TweetFields).
		SetQueryParam("user.fields", UserFields).
		SetQueryParam("media.fields", MediaFields).
		SetQueryParam("expansions", Expansions).
		Get("/tweets/{id}")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if resp.IsError() {
		return nil, errors.Errorf("request failed with status: %d - Body: %s", resp.StatusCode(), resp.Body())
	}

	tweet := transformTweetResponses(response)
	return &tweet, nil
}

func (t *v2Impl) AddUserToList(ctx context.Context, accessToken AccessToken, listID string, param AddUserToListParam) error {
	var response APIV2Response[any]
	resp, err := t.httpClient.R().
		SetContext(ctx).
		SetResult(&response).
		SetAuthToken(string(accessToken)).
		SetBody(param).
		SetPathParam("id", listID).
		Post("/lists/{id}/members")
	if err != nil {
		return errors.WithStack(err)
	}

	if resp.IsError() {
		return errors.Errorf("request failed with status: %d - Body: %s", resp.StatusCode(), resp.Body())
	}

	return nil
}

func (t *v2Impl) GetListMembers(ctx context.Context, accessToken AccessToken, listID string, param GetListMembersParam) ([]User, Meta, error) {
	var result APIV2Response[[]User]
	req := t.httpClient.
		R().
		SetContext(ctx).
		SetAuthToken(string(accessToken)).
		SetPathParam("id", listID).
		SetQueryParam("max_results", fmt.Sprintf("%d", param.Limit)).
		SetQueryParam("tweet.fields", TweetFields).
		SetQueryParam("user.fields", UserFields).
		SetResult(&result)

	if param.PaginationToken != "" {
		req.SetQueryParam("pagination_token", param.PaginationToken)
	}

	resp, err := req.Get("/lists/{id}/members")
	if err != nil {
		return nil, Meta{}, err
	}
	if resp.IsError() {
		return nil, Meta{}, errors.Errorf("Response code: %d - Body: %s", resp.StatusCode(), resp.String())
	}

	return kit.Map(result.Data, transformUser), lo.FromPtr(result.Meta), nil
}

func (t *v2Impl) GetTweets(ctx context.Context, accessToken AccessToken, tweetIDs []string) ([]Tweet, error) {
	var response APIV2Response[[]Tweet]
	resp, err := t.httpClient.R().
		SetContext(ctx).
		SetResult(&response).
		SetAuthToken(string(accessToken)).
		SetQueryParam("ids", strings.Join(tweetIDs, ",")).
		SetQueryParam("tweet.fields", TweetFields).
		SetQueryParam("user.fields", UserFields).
		SetQueryParam("media.fields", MediaFields).
		SetQueryParam("expansions", Expansions).
		Get("/tweets")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if resp.IsError() {
		return nil, errors.Errorf("request failed with status: %d - Body: %s", resp.StatusCode(), resp.Body())
	}

	return transformTweetsResponses(response), nil
}

func (t *v2Impl) GetMentions(ctx context.Context, accessToken AccessToken, userID string, param FilterTweetsParam) ([]Tweet, Meta, error) {
	var response APIV2Response[[]Tweet]
	req := t.httpClient.R().
		SetContext(ctx).
		SetAuthToken(string(accessToken)).
		SetResult(&response).
		SetPathParam("id", userID).
		SetQueryParam("max_results", fmt.Sprintf("%d", param.Limit)).
		SetQueryParam("tweet.fields", TweetFields).
		SetQueryParam("user.fields", UserFields).
		SetQueryParam("media.fields", MediaFields).
		SetQueryParam("expansions", Expansions)

	if param.SinceID != "" {
		req.SetQueryParam("since_id", param.SinceID)
	} else if param.StartTime != "" {
		req.SetQueryParam("start_time", param.StartTime)
	}

	if param.UntilID != "" {
		req.SetQueryParam("until_id", param.UntilID)
	} else if param.EndTime != "" {
		req.SetQueryParam("end_time", param.EndTime)
	}

	if param.PaginationToken != "" {
		req.SetQueryParam("pagination_token", param.PaginationToken)
	}

	resp, err := req.Get("/users/{id}/mentions")
	if err != nil {
		return nil, Meta{}, errors.WithStack(err)
	}

	if resp.IsError() {
		return nil, Meta{}, errors.Errorf("request failed with status: %d - Body: %s", resp.StatusCode(), resp.Body())
	}

	return transformTweetsResponses(response), lo.FromPtr(response.Meta), nil
}

func (t *v2Impl) SearchRecentTweets(ctx context.Context, accessToken AccessToken, param GetRecentTweetsParam) ([]Tweet, Meta, error) {
	var response APIV2Response[[]Tweet]
	req := t.httpClient.R().
		SetContext(ctx).
		SetAuthToken(string(accessToken)).
		SetResult(&response).
		SetQueryParam("query", param.Query).
		SetQueryParam("max_results", fmt.Sprintf("%d", param.Limit)).
		SetQueryParam("tweet.fields", TweetFields).
		SetQueryParam("user.fields", UserFields).
		SetQueryParam("media.fields", MediaFields).
		SetQueryParam("expansions", Expansions)

	if param.SinceID != "" {
		req.SetQueryParam("since_id", param.SinceID)
	} else if param.StartTime != "" {
		req.SetQueryParam("start_time", param.StartTime)
	}

	if param.UntilID != "" {
		req.SetQueryParam("until_id", param.UntilID)
	} else if param.EndTime != "" {
		req.SetQueryParam("end_time", param.EndTime)
	}

	if param.PaginationToken != "" {
		req.SetQueryParam("pagination_token", param.PaginationToken)
	}

	resp, err := req.Get("/tweets/search/recent")
	if err != nil {
		return nil, Meta{}, errors.WithStack(err)
	}

	if resp.IsError() {
		return nil, Meta{}, errors.Errorf("request failed with status: %d - Body: %s", resp.StatusCode(), resp.Body())
	}

	return transformTweetsResponses(response), lo.FromPtr(response.Meta), nil
}

func transformTweetsResponses(response APIV2Response[[]Tweet]) []Tweet {
	userByID := lo.KeyBy(response.Includes.Users, func(item User) string {
		return item.ID
	})
	mediaByID := lo.KeyBy(response.Includes.Media, func(item Media) string {
		return item.MediaKey
	})

	return kit.Map(response.Data, func(item Tweet) Tweet {
		item.Author = transformUser(userByID[item.AuthorID])
		item.Medias = kit.Map(item.Attachments.MediaKeys, func(mediaKey string) Media {
			return mediaByID[mediaKey]
		})
		return item
	})
}

func transformTweetResponses(response APIV2Response[Tweet]) Tweet {
	userByID := lo.KeyBy(response.Includes.Users, func(item User) string {
		return item.ID
	})
	mediaByID := lo.KeyBy(response.Includes.Media, func(item Media) string {
		return item.MediaKey
	})

	item := response.Data

	item.Author = transformUser(userByID[item.AuthorID])
	item.Medias = kit.Map(item.Attachments.MediaKeys, func(mediaKey string) Media {
		return mediaByID[mediaKey]
	})
	return item
}

func transformUser(user User) User {
	user.ProfileImageURL = strings.ReplaceAll(user.ProfileImageURL, "_normal.", ".")
	return user
}
