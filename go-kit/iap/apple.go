package iap

import (
	"context"

	"github.com/awa/go-iap/appstore"
	"github.com/awa/go-iap/appstore/api"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

var _ SubscriptionProvider = (*AppleProvider)(nil)

type AppleProvider struct {
	apiClientProduction *api.StoreClient
	apiClientSandbox    *api.StoreClient
	appstoreClient      *appstore.Client
}

func NewAppleProvider(cfg AppleConfig) *AppleProvider {
	return &AppleProvider{
		apiClientProduction: api.NewStoreClient(&api.StoreConfig{
			KeyContent: []byte(cfg.AccountPrivateKey), // Loads a .p8 certificate
			KeyID:      cfg.KeyID,                     // Your private key ID from App Store Connect (Ex: 2X9R4HXF34)
			BundleID:   cfg.BundleID,                  // Your app’s bundle ID
			Issuer:     cfg.Issuer,                    // Your issuer ID from the Keys page in App Store Connect (Ex: "57246542-96fe-1a63-e053-0824d011072a")
			Sandbox:    false,
		}),
		apiClientSandbox: api.NewStoreClient(&api.StoreConfig{
			KeyContent: []byte(cfg.AccountPrivateKey), // Loads a .p8 certificate
			KeyID:      cfg.KeyID,                     // Your private key ID from App Store Connect (Ex: 2X9R4HXF34)
			BundleID:   cfg.BundleID,                  // Your app’s bundle ID
			Issuer:     cfg.Issuer,                    // Your issuer ID from the Keys page in App Store Connect (Ex: "57246542-96fe-1a63-e053-0824d011072a")
			Sandbox:    true,
		}),
		appstoreClient: appstore.New(),
	}
}

func (a *AppleProvider) VerifyEventAuth(_ context.Context, _ string) error {
	// TODO: verify jwt token
	return nil
}

func (a *AppleProvider) ParseEvent(_ context.Context, requestBody map[string]any) (*Notification, error) {
	tokenStr, found := requestBody["signedPayload"]
	if !found {
		return nil, errors.Errorf("missing signedPayload")
	}
	var token jwt.Token
	err := a.appstoreClient.ParseNotificationV2(tokenStr.(string), &token)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.Errorf("invalid claims")
	}

	notificationType, found := claims["notificationType"]
	if !found {
		return nil, errors.Errorf("missing notification type")
	}

	// TODO: parse claims

	return &Notification{
		Raw:           requestBody,
		TransactionID: "",
		EventType:     EventType(notificationType.(string)),
	}, nil
}

func (a *AppleProvider) GetTransactionInfo(ctx context.Context, _, transactionID string) (*Transaction, error) {
	transactionInfo, err := a.apiClientProduction.GetTransactionInfo(ctx, transactionID)
	if err != nil {
		if !errors.Is(err, api.TransactionIdNotFoundError) {
			return nil, errors.Wrapf(err, "AppleProvider.GetTransactionInfo-Prod: %s", transactionID)
		}

		transactionInfo, err = a.apiClientSandbox.GetTransactionInfo(ctx, transactionID)
		if err != nil {
			return nil, errors.Wrapf(err, "AppleProvider.GetTransactionInfo-Sandbox: %s", transactionID)
		}
	}

	signedTransaction, err := a.apiClientProduction.ParseSignedTransaction(transactionInfo.SignedTransactionInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "AppleProvider.ParseSignedTransaction")
	}

	return &Transaction{
		Raw:                   signedTransaction,
		OriginalTransactionID: signedTransaction.OriginalTransactionId,
		OriginalPurchaseDate:  DateTime(signedTransaction.OriginalPurchaseDate),
		TransactionID:         signedTransaction.TransactionID,
		PurchaseDate:          DateTime(signedTransaction.PurchaseDate),
		PurchaseState:         PurchaseStatePurchased,
		BundleID:              signedTransaction.BundleID,
		ProductID:             signedTransaction.ProductID,
		ExpiresDate:           DateTime(signedTransaction.ExpiresDate),
		Quantity:              int64(signedTransaction.Quantity),
		Type:                  ProductType(signedTransaction.Type),
		ReferenceID:           signedTransaction.AppAccountToken,
		SignedDate:            DateTime(signedTransaction.SignedDate),
		OfferType:             signedTransaction.OfferType,
		OfferIdentifier:       signedTransaction.OfferIdentifier,
		RevocationDate:        DateTime(signedTransaction.RevocationDate),
		RevocationReason:      signedTransaction.RevocationReason,
		IsUpgraded:            signedTransaction.IsUpgraded,
		Storefront:            signedTransaction.Storefront,
		StorefrontID:          signedTransaction.StorefrontId,
		TransactionReason:     string(signedTransaction.TransactionReason),
		IsSandbox:             signedTransaction.Environment == api.Sandbox,
	}, nil
}

func (a *AppleProvider) AcknowledgeSubscription(_ context.Context, _, _, _ string) error {
	// nothing to do
	return nil
}
