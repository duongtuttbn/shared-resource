package evm

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"github.com/pkg/errors"
)

type Wallet struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

func NewWalletFromHex(privateKeyHex string) (*Wallet, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, errors.Wrap(err, "crypto.HexToECDSA")
	}

	return NewWallet(privateKey)
}

func NewWallet(privateKey *ecdsa.PrivateKey) (*Wallet, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	return &Wallet{
		privateKey: privateKey,
		address:    crypto.PubkeyToAddress(*publicKeyECDSA),
	}, nil
}

func (w *Wallet) Address() common.Address {
	return w.address
}

func (w *Wallet) SignTx(_ context.Context, tx *types.Transaction) (*types.Transaction, error) {
	return types.SignTx(tx, types.NewLondonSigner(tx.ChainId()), w.privateKey)
}

func (w *Wallet) SignTxInChain(chainID *big.Int) func(*types.Transaction) (*types.Transaction, error) {
	return func(tx *types.Transaction) (*types.Transaction, error) {
		return types.SignTx(tx, types.NewLondonSigner(chainID), w.privateKey)
	}
}

func (w *Wallet) SendNativeToken(ctx context.Context, client *Client, to common.Address, amount *big.Int, gasTipCap ...int64) (common.Hash, error) {
	tx, err := BuildTx(ctx, client, w.Address(), to, amount, nil, gasTipCap...)
	if err != nil {
		return common.Hash{}, err
	}

	return w.SendTx(ctx, client, tx)
}

func (w *Wallet) SendERC20Token(ctx context.Context, client *Client, tokenAddress, to common.Address, amount *big.Int, gasTipCap ...int64) (common.Hash, error) {
	tx, err := BuildTransferERC20Tx(ctx, client, tokenAddress, w.Address(), to, amount, gasTipCap...)
	if err != nil {
		return common.Hash{}, err
	}

	return w.SendTx(ctx, client, tx)
}

func (w *Wallet) SendTx(ctx context.Context, client *Client, tx *types.Transaction) (common.Hash, error) {
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return common.Hash{}, errors.Wrap(err, "client.ChainID")
	}

	signedTx, err := w.SignTxInChain(chainID)(tx)
	if err != nil {
		return common.Hash{}, errors.Wrap(err, "wallet.SignTxInChain")
	}

	return SendSignedTx(ctx, client, signedTx)
}

func (w *Wallet) Sign(payload ...any) ([]byte, error) {
	messageHash := solsha3.SoliditySHA3(payload...)

	signatureBytes, err := crypto.Sign(accounts.TextHash(messageHash), w.privateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "sign message")
	}

	// We need this to correct v = 0,1 to v = 27,28 - or else all will break
	if signatureBytes[64] == 0 || signatureBytes[64] == 1 {
		signatureBytes[64] += 27
	}

	return signatureBytes, nil
}
