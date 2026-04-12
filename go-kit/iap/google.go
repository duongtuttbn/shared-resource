package iap

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"
	"github.com/duongtuttbn/shared-resource/go-kit/dt"
	"github.com/duongtuttbn/shared-resource/go-kit/kit"

	"github.com/awa/go-iap/playstore"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/idtoken"
)

var _ SubscriptionProvider = (*GoogleProvider)(nil)

type GoogleProvider struct {
	client   *playstore.Client
	clientID string
}

func NewGoogleProvider(cfg GoogleConfig) (*GoogleProvider, error) {
	jsonKey := []byte(cfg.JSONKey)
	var serviceAccountMap dt.Map
	err := json.Unmarshal(jsonKey, &serviceAccountMap)
	if err != nil {
		return nil, err
	}

	client, err := playstore.New(jsonKey)
	if err != nil {
		return nil, err
	}
	return &GoogleProvider{
		client:   client,
		clientID: serviceAccountMap["client_id"].(string),
	}, nil
}

type realTimeDeveloperNotification struct {
	Message struct {
		Attributes map[string]any `json:"attributes"`
		Data       string         `json:"data" description:"base64 encoded"`
		MessageID  string         `json:"messageId"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

var googleSubscriptionNotificationMap = map[playstore.SubscriptionNotificationType]EventType{
	playstore.SubscriptionNotificationTypeRecovered:            EventTypeRecovered,
	playstore.SubscriptionNotificationTypeRenewed:              EventTypeRenewed,
	playstore.SubscriptionNotificationTypeCanceled:             EventTypeCanceled,
	playstore.SubscriptionNotificationTypePurchased:            EventTypePurchased,
	playstore.SubscriptionNotificationTypeAccountHold:          EventTypeAccountHold,
	playstore.SubscriptionNotificationTypeGracePeriod:          EventTypeGracePeriod,
	playstore.SubscriptionNotificationTypeRestarted:            EventTypeRestarted,
	playstore.SubscriptionNotificationTypePriceChangeConfirmed: EventTypePriceChangeConfirmed,
	playstore.SubscriptionNotificationTypeDeferred:             EventTypeDeferred,
	playstore.SubscriptionNotificationTypePaused:               EventTypePaused,
	playstore.SubscriptionNotificationTypePauseScheduleChanged: EventTypePauseScheduleChanged,
	playstore.SubscriptionNotificationTypeRevoked:              EventTypeRevoked,
	playstore.SubscriptionNotificationTypeExpired:              EventTypeExpired,
}

func (a *GoogleProvider) VerifyEventAuth(ctx context.Context, jwtToken string) error {
	payload, err := idtoken.Validate(ctx, jwtToken, "")
	if err != nil {
		return errors.Wrapf(err, "GoogleProvider.VerifyEventAuth")
	}

	if payload.Subject != a.clientID {
		return errors.New("invalid subject")
	}
	return nil
}

func (a *GoogleProvider) ParseEvent(_ context.Context, requestBody map[string]any) (*Notification, error) {
	notification, err := kit.ConvertType[realTimeDeveloperNotification](requestBody)
	if err != nil {
		return nil, errors.Wrapf(err, "ConvertToType requestBody")
	}
	bytes, err := base64.StdEncoding.DecodeString(notification.Message.Data)
	if err != nil {
		return nil, errors.Wrapf(err, "DecodeString")
	}

	var developerNotification playstore.DeveloperNotification

	err = json.Unmarshal(bytes, &developerNotification)
	if err != nil {
		return nil, errors.Wrapf(err, "Unmarshal")
	}

	eventType, ok := googleSubscriptionNotificationMap[developerNotification.SubscriptionNotification.NotificationType]
	if !ok {
		eventType = "unknown"
	}

	millis, err := strconv.ParseInt(developerNotification.EventTimeMillis, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid event time")
	}
	return &Notification{
		Raw:           notification,
		PackageName:   developerNotification.PackageName,
		TransactionID: developerNotification.SubscriptionNotification.PurchaseToken,
		EventType:     eventType,
		EventTime:     time.UnixMilli(millis),
	}, nil
}

func (a *GoogleProvider) GetTransactionInfo(ctx context.Context, packageName, purchaseToken string) (*Transaction, error) {
	sub, err := a.client.VerifySubscriptionV2(ctx, packageName, purchaseToken)
	if err != nil {
		return nil, errors.Wrapf(err, "GoogleProvider.VerifyProduct")
	}

	rawPurchaseState := sub.SubscriptionState
	var purchaseState PurchaseState
	switch rawPurchaseState {
	case "SUBSCRIPTION_STATE_ACTIVE":
		purchaseState = PurchaseStatePurchased
	case "SUBSCRIPTION_STATE_PENDING":
		purchaseState = PurchaseStatePending
	default:
		purchaseState = PurchaseStateCanceled
	}

	if len(sub.LineItems) == 0 {
		return nil, errors.New("googleiap: no line items")
	}

	item := sub.LineItems[0]

	return &Transaction{
		Raw:                   sub,
		OriginalTransactionID: sub.LinkedPurchaseToken,
		OriginalPurchaseDate:  0,
		TransactionID:         purchaseToken,
		PurchaseDate:          a.parseTime(sub.StartTime),
		RawPurchaseState:      rawPurchaseState,
		PurchaseState:         purchaseState,
		BundleID:              packageName,
		ProductID:             item.ProductId,
		ExpiresDate:           a.parseTime(item.ExpiryTime),
		ReferenceID:           lo.FromPtr(sub.ExternalAccountIdentifiers).ObfuscatedExternalAccountId,
		IsUpgraded:            sub.LinkedPurchaseToken != "",
		IsSandbox:             sub.TestPurchase != nil,
	}, nil
}

func (a *GoogleProvider) AcknowledgeSubscription(ctx context.Context, packageName, purchaseToken string, developerPayload string) error {
	return a.client.AcknowledgeSubscription(ctx, packageName, "", purchaseToken, &androidpublisher.SubscriptionPurchasesAcknowledgeRequest{
		DeveloperPayload: developerPayload,
	})
}

func (a *GoogleProvider) parseTime(timeStr string) DateTime {
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err == nil {
		return DateTime(t.UnixMilli())
	}
	return 0
}
