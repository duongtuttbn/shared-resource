package twitterapi

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"
)

const (
	defaultRetryWaitTime = 15 * time.Second
	baseURL              = "https://api.x.com/2"

	UserFields  = "created_at,description,entities,id,location,name,pinned_tweet_id,profile_image_url,protected,public_metrics,url,username,verified,verified_type,withheld"
	TweetFields = "attachments,author_id,context_annotations,conversation_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,public_metrics,referenced_tweets,reply_settings,source,text,withheld,edit_history_tweet_ids"
	MediaFields = "duration_ms,height,media_key,preview_image_url,public_metrics,type,url,width"
	Expansions  = "author_id,attachments.media_keys"

	ScopeTweetRead          Scope = "tweet.read"
	ScopeTweetWrite         Scope = "tweet.write"
	ScopeTweetModerateWrite Scope = "tweet.moderate.write"
	ScopeUsersEmail         Scope = "users.email"
	ScopeUsersRead          Scope = "users.read"
	ScopeFollowsRead        Scope = "follows.read"
	ScopeFollowsWrite       Scope = "follows.write"
	ScopeOfflineAccess      Scope = "offline.access"
	ScopeSpaceRead          Scope = "space.read"
	ScopeMuteRead           Scope = "mute.read"
	ScopeMuteWrite          Scope = "mute.write"
	ScopeLikeRead           Scope = "like.read"
	ScopeLikeWrite          Scope = "like.write"
	ScopeListRead           Scope = "list.read"
	ScopeListWrite          Scope = "list.write"
	ScopeBlockRead          Scope = "block.read"
	ScopeBlockWrite         Scope = "block.write"
	ScopeBookmarkRead       Scope = "bookmark.read"
	ScopeBookmarkWrite      Scope = "bookmark.write"
	ScopeMediaWrite         Scope = "media.write"

	ChallengeMethodS256  ChallengeMethod = "S256"
	ChallengeMethodPlain ChallengeMethod = "plain"
)

type (
	Scope           string
	ChallengeMethod string
	AccessToken     string
	ClientID        string
)

type V2 interface {
	GetUserInfo(ctx context.Context, accessToken AccessToken) (*User, error)
	GetUserByID(ctx context.Context, accessToken AccessToken, id string) (*User, error)
	GetUserByIDs(ctx context.Context, accessToken AccessToken, ids []string) ([]User, error)
	GetMentions(ctx context.Context, accessToken AccessToken, userID string, param FilterTweetsParam) ([]Tweet, Meta, error)
	GetTweets(ctx context.Context, accessToken AccessToken, ids []string) ([]Tweet, error)
	SearchRecentTweets(ctx context.Context, accessToken AccessToken, param GetRecentTweetsParam) ([]Tweet, Meta, error)
	GetTweet(ctx context.Context, accessToken AccessToken, id string) (*Tweet, error)
	CreateTweet(ctx context.Context, accessToken AccessToken, param CreateTweetParam) (*SimpleTweet, error)
	AddUserToList(ctx context.Context, accessToken AccessToken, listID string, param AddUserToListParam) error
	GetListMembers(ctx context.Context, accessToken AccessToken, listID string, param GetListMembersParam) ([]User, Meta, error)
}

type V2Oauth interface {
	V2
	GetAccessToken(ctx context.Context, code string, codeVerifier string) (*TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	GetOAuthURL(state string, codeChallenge string, codeChallengeMethod ChallengeMethod, additionalScopes ...Scope) string
}

type Config struct {
	RedirectURI  string   `json:"redirect_uri" mapstructure:"redirect_uri"`
	ClientID     ClientID `json:"client_id" mapstructure:"client_id"`
	ClientSecret string   `json:"client_secret" mapstructure:"client_secret"`
}

type APIV2Response[T any] struct {
	Data     T       `json:"data"`
	Errors   []Error `json:"errors"`
	Meta     *Meta   `json:"meta"`
	Includes struct {
		Users []User  `json:"users"`
		Media []Media `json:"media"`
	}
}

type FilterTweetsParam struct {
	Limit           int    `json:"limit"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	SinceID         string `json:"since_id"`
	UntilID         string `json:"until_id"`
	PaginationToken string `json:"pagination_token"`
}

type GetRecentTweetsParam struct {
	FilterTweetsParam
	Query string `json:"query"`
}

type User struct {
	PublicMetrics struct {
		FollowersCount int64 `json:"followers_count"`
		FollowingCount int64 `json:"following_count"`
		TweetCount     int64 `json:"tweet_count"`
		ListedCount    int64 `json:"listed_count"`
		LikeCount      int64 `json:"like_count"`
		MediaCount     int64 `json:"media_count"`
	} `json:"public_metrics"`
	ProfileImageURL string    `json:"profile_image_url"`
	Name            string    `json:"name"`
	Verified        bool      `json:"verified"`
	VerifiedType    string    `json:"verified_type"`
	Username        string    `json:"username"`
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	Protected       bool      `json:"protected"`
	Description     string    `json:"description"`
}

func (s User) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *User) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), s)
}

type CreateTweetParam struct {
	Text  string `json:"text"`
	Reply *Reply `json:"reply,omitempty"`
}

type Reply struct {
	InReplyToTweetID string `json:"in_reply_to_tweet_id"`
}

type AddUserToListParam struct {
	UserID string `json:"user_id"`
}

type GetListMembersParam struct {
	Limit           int    `json:"limit"`
	PaginationToken string `json:"pagination_token"`
}

type Meta struct {
	ResultCount int    `json:"result_count"`
	NextToken   string `json:"next_token"`
}

type Error struct {
	Parameter    string `json:"parameter"`
	ResourceID   string `json:"resource_id"`
	Value        string `json:"value"`
	Detail       string `json:"detail"`
	Title        string `json:"title"`
	ResourceType string `json:"resource_type"`
	Type         string `json:"type"`
}

type SimpleTweet struct {
	EditHistoryTweetIDs []string `json:"edit_history_tweet_ids"`
	ID                  string   `json:"id"`
	Text                string   `json:"text"`
}

type MentionUser struct {
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Username string `json:"username"`
	ID       string `json:"id"`
}

type EntityURL struct {
	Start       int    `json:"start"`
	End         int    `json:"end"`
	URL         string `json:"url"`
	ExpandedURL string `json:"expanded_url"`
	DisplayURL  string `json:"display_url"`
}

type ReferencedTweet struct {
	Type ReferencedType `json:"type"`
	ID   string         `json:"id"`
}

type Media struct {
	Type            string `json:"type"`
	URL             string `json:"url"`
	Width           int    `json:"width"`
	MediaKey        string `json:"media_key"`
	Height          int    `json:"height"`
	PreviewImageURL string `json:"preview_image_url"`
	DurationMs      int    `json:"duration_ms"`
	PublicMetrics   struct {
		ViewCount int `json:"view_count"`
	} `json:"public_metrics"`
}

type (
	TweetType      string
	ReferencedType string
)

const (
	TweetTypeOriginal TweetType = "original"
	TweetTypeRetweet  TweetType = "retweet"
	TweetTypeQuote    TweetType = "quote"
	TweetTypeReply    TweetType = "reply"

	ReferencedTypeQuoted    ReferencedType = "quoted"
	ReferencedTypeRetweeted ReferencedType = "retweeted"
	ReferencedTypeRepliedTo ReferencedType = "replied_to"
)

var tweetTypeMap = map[ReferencedType]TweetType{
	ReferencedTypeQuoted:    TweetTypeQuote,
	ReferencedTypeRetweeted: TweetTypeRetweet,
	ReferencedTypeRepliedTo: TweetTypeReply,
}

type Tweet struct {
	SimpleTweet
	Entities struct {
		Urls     []EntityURL   `json:"urls"`
		Mentions []MentionUser `json:"mentions"`
	} `json:"entities"`
	Attachments struct {
		MediaKeys []string `json:"media_keys"`
	} `json:"attachments"`
	AuthorID          string             `json:"author_id"`
	PossiblySensitive bool               `json:"possibly_sensitive"`
	Lang              string             `json:"lang"`
	ConversationID    string             `json:"conversation_id"`
	ReplySettings     string             `json:"reply_settings"`
	ReferencedTweets  []ReferencedTweet  `json:"referenced_tweets"`
	PublicMetrics     TweetPublicMetrics `json:"public_metrics"`
	CreatedAt         time.Time          `json:"created_at"`
	Author            User               `json:"author"`
	Medias            []Media            `json:"medias"`
}

func (s Tweet) GetType() TweetType {
	for _, referencedTweet := range s.ReferencedTweets {
		tweetType, found := tweetTypeMap[referencedTweet.Type]
		if found {
			return tweetType
		}
	}

	return TweetTypeOriginal
}

func (s Tweet) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *Tweet) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), s)
}

type TweetPublicMetrics struct {
	RetweetCount    int `json:"retweet_count"`
	ReplyCount      int `json:"reply_count"`
	LikeCount       int `json:"like_count"`
	QuoteCount      int `json:"quote_count"`
	BookmarkCount   int `json:"bookmark_count"`
	ImpressionCount int `json:"impression_count"`
}
type TokenResponse struct {
	TokenType    string      `json:"token_type"`
	ExpiresIn    int         `json:"expires_in"`
	AccessToken  AccessToken `json:"access_token"`
	Scope        string      `json:"scope"`
	RefreshToken string      `json:"refresh_token"`
}
