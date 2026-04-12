package iap

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"
)

type (
	GoogleConfig struct {
		JSONKey string `json:"json_key"`
	}
	AppleConfig struct {
		AccountPrivateKey string `json:"account_private_key"`
		KeyID             string `json:"key_id"`
		BundleID          string `json:"bundle_id"`
		Issuer            string `json:"issuer"`
	}

	SubscriptionProvider interface {
		VerifyEventAuth(ctx context.Context, token string) error
		ParseEvent(ctx context.Context, requestBody map[string]any) (*Notification, error)
		GetTransactionInfo(ctx context.Context, packageName string, transactionID string) (*Transaction, error)
		AcknowledgeSubscription(ctx context.Context, packageName string, transactionID string, developerPayload string) error
	}

	Platform          string
	ProductType       string
	TransactionStatus string
	PurchaseState     string
	EventType         string
)

const (
	PlatformApple  Platform = "apple"
	PlatformGoogle Platform = "google"

	ProductTypeOneTime      ProductType = "one_time"
	ProductTypeSubscription ProductType = "subscription"

	TransactionStatusCreated   TransactionStatus = "created"
	TransactionStatusPurchased TransactionStatus = "purchased"
	TransactionStatusCanceled  TransactionStatus = "canceled"

	PurchaseStatePending   PurchaseState = "pending"
	PurchaseStatePurchased PurchaseState = "purchased"
	PurchaseStateCanceled  PurchaseState = "canceled"

	EventTypeRecovered            EventType = "recovered"
	EventTypeRenewed              EventType = "renewed"
	EventTypeCanceled             EventType = "canceled"
	EventTypePurchased            EventType = "purchased"
	EventTypeAccountHold          EventType = "account_hold"
	EventTypeGracePeriod          EventType = "grace_period"
	EventTypeRestarted            EventType = "restarted"
	EventTypePriceChangeConfirmed EventType = "price_change_confirmed"
	EventTypeDeferred             EventType = "deferred"
	EventTypePaused               EventType = "paused"
	EventTypePauseScheduleChanged EventType = "pause_schedule_changed"
	EventTypeRevoked              EventType = "revoked"
	EventTypeExpired              EventType = "expired"
)

type DateTime int64 // Unix milliseconds

func (d DateTime) ToTime() *time.Time {
	if d == 0 {
		return nil
	}
	t := time.UnixMilli(int64(d))
	return &t
}

type Notification struct {
	Raw           any       `json:"raw"`
	EventTime     time.Time `json:"event_time"`
	PackageName   string    `json:"package_name"`
	TransactionID string    `json:"transaction_id"`
	EventType     EventType `json:"event_type"`
}

func (s Notification) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *Notification) Scan(value any) error {
	return json.Unmarshal(value.([]byte), s)
}

type Transaction struct {
	Raw                   any           `json:"raw"`
	OriginalTransactionID string        `json:"original_transaction_id"`
	OriginalPurchaseDate  DateTime      `json:"original_purchase_date"`
	TransactionID         string        `json:"transaction_id"`
	PurchaseDate          DateTime      `json:"purchase_date"`
	PurchaseState         PurchaseState `json:"purchase_state"`
	RawPurchaseState      string        `json:"raw_purchase_state"`
	BundleID              string        `json:"bundle_id"`
	ProductID             string        `json:"product_id"`
	ExpiresDate           DateTime      `json:"expires_date"`
	Quantity              int64         `json:"quantity"`
	Type                  ProductType   `json:"type"`
	ReferenceID           string        `json:"reference_id"`
	SignedDate            DateTime      `json:"signed_date"`
	OfferType             int32         `json:"offer_type"`
	OfferIdentifier       string        `json:"offer_identifier"`
	RevocationDate        DateTime      `json:"revocation_date"`
	RevocationReason      *int32        `json:"revocation_reason"`
	IsUpgraded            bool          `json:"is_upgraded"`
	Storefront            string        `json:"storefront"`
	StorefrontID          string        `json:"storefront_id"`
	TransactionReason     string        `json:"transaction_reason"`
	IsSandbox             bool          `json:"is_sandbox"`
}
