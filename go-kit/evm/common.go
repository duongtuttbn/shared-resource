package evm

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

func SendSignedTx(ctx context.Context, client *Client, signedTx *types.Transaction) (common.Hash, error) {
	if err := client.SendTransaction(ctx, signedTx); err != nil {
		if strings.Contains(err.Error(), "insufficient funds") {
			return common.Hash{}, ErrInsufficientBalance
		}

		return common.Hash{}, errors.Wrap(err, "client.SendTransaction")
	}

	return signedTx.Hash(), nil
}

func BuildTx(ctx context.Context, client *Client, from, to common.Address, value *big.Int, data []byte, gasTipCap ...int64) (*types.Transaction, error) {
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: value,
		Data:  data,
	})
	if err != nil {
		return nil, errors.Wrap(err, "client.EstimateGas")
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "client.ChainID")
	}

	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, errors.Wrap(err, "client.PendingNonceAt")
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "client.SuggestGasPrice")
	}

	tipCap := big.NewInt(0)
	if len(gasTipCap) > 0 {
		tipCap.SetInt64(gasTipCap[0])
	} else {
		suggestTipCap, err := client.SuggestGasTipCap(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "client.SuggestGasTipCap")
		}
		if suggestTipCap.Cmp(big.NewInt(0)) > 0 {
			tipCap = suggestTipCap
		}
	}

	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: gasPrice,
		Gas:       gasLimit,
		To:        &to,
		Value:     value,
		Data:      data,
	}), nil
}

func BuildTransferERC20Tx(ctx context.Context, client *Client, tokenAddress, from, to common.Address, amount *big.Int, gasTipCap ...int64) (*types.Transaction, error) {
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]

	paddedAddress := common.LeftPadBytes(to.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	return BuildTx(ctx, client, from, tokenAddress, big.NewInt(0), data, gasTipCap...)
}
