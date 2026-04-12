package evm

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/crypto/sha3"
	"github.com/duongtuttbn/shared-resource/go-kit/dt"
)

type Service interface {
	GetAllowance(ctx context.Context, tokenAddress Address, ownerAddress, spenderAddress Address) (*dt.NumericInt, error)
	GetBalance(ctx context.Context, tokenAddress, ownerAddress Address) (*dt.NumericInt, int, error)
	GetTokenInfo(ctx context.Context, tokenAddress Address) (*TokenInfo, error)
	BuildApproveToken(ctx context.Context, tokenAddress Address, ownerAddress, spenderAddress Address, amount *dt.NumericInt, gasTipCap ...int64) (*types.Transaction, error)
	BuildTransferNativeToken(ctx context.Context, from, to Address, amount *dt.NumericInt, gasTipCap ...int64) (*types.Transaction, error)
	BuildTransferToken(ctx context.Context, mint, from, to Address, amount *dt.NumericInt, gasTipCap ...int64) (*types.Transaction, error)
	BuildTx(ctx context.Context, from, to Address, value *dt.NumericInt, data []byte, gasTipCap ...int64) (*types.Transaction, error)
	SendTransaction(ctx context.Context, transaction *types.Transaction) (TxHash, error)
	WaitTxAndGetReceipt(ctx context.Context, txHash TxHash, timeout ...time.Duration) (*types.Receipt, error)
}

type serviceImpl struct {
	clientPool *ClientPool
}

func NewService(clientPool *ClientPool) Service {
	return &serviceImpl{
		clientPool: clientPool,
	}
}

func (s *serviceImpl) GetBalance(ctx context.Context, tokenAddress, ownerAddress Address) (*dt.NumericInt, int, error) {
	if NativeTokenAddress.Equals(tokenAddress) {
		balance, err := s.clientPool.GetClient().BalanceAt(ctx, ownerAddress.ToETHAddress(), nil)
		if err != nil {
			return nil, 0, err
		}

		return dt.NumFromBigInt(balance), 18, nil
	}

	tokenInfo, err := s.GetTokenInfo(ctx, tokenAddress)
	if err != nil {
		return nil, 0, err
	}

	balance, err := s.clientPool.GetClient().GetERC20Balance(
		ctx,
		tokenAddress.ToETHAddress(),
		ownerAddress.ToETHAddress(),
	)
	if err != nil {
		return nil, 0, errors.Wrap(err, "get balance")
	}

	return dt.NumFromBigInt(balance), tokenInfo.Decimals, nil
}

func (s *serviceImpl) SendTransaction(ctx context.Context, transaction *types.Transaction) (TxHash, error) {
	txHash, err := SendSignedTx(ctx, s.clientPool.GetClient(), transaction)
	if err != nil {
		return "", err
	}

	return TxHash(txHash.String()), nil
}

func (s *serviceImpl) GetTokenInfo(ctx context.Context, tokenAddress Address) (*TokenInfo, error) {
	if tokenAddress.Equals(NativeTokenAddress) {
		return &TokenInfo{Address: tokenAddress, Decimals: 18}, nil
	}
	decimalsFnSignature := []byte("decimals()")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(decimalsFnSignature)
	methodID := hash.Sum(nil)[:4]

	decimals, err := s.clientPool.GetClient().CallContract(ctx, ethereum.CallMsg{
		To:   lo.ToPtr(tokenAddress.ToETHAddress()),
		Data: methodID,
	}, nil)
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		Address:  tokenAddress,
		Decimals: int(big.NewInt(0).SetBytes(decimals).Int64()),
	}, nil
}

func (s *serviceImpl) BuildTransferNativeToken(ctx context.Context, from, to Address, amount *dt.NumericInt, gasTipCap ...int64) (*types.Transaction, error) {
	client := s.clientPool.GetClient()
	return BuildTx(ctx, client, from.ToETHAddress(), to.ToETHAddress(), amount.ToBigInt(), nil, gasTipCap...)
}

func (s *serviceImpl) BuildTx(ctx context.Context, from, to Address, value *dt.NumericInt, data []byte, gasTipCap ...int64) (*types.Transaction, error) {
	client := s.clientPool.GetClient()
	return BuildTx(ctx, client, from.ToETHAddress(), to.ToETHAddress(), value.ToBigInt(), data, gasTipCap...)
}

func (s *serviceImpl) BuildTransferToken(ctx context.Context, tokenAddress Address, from, to Address, amount *dt.NumericInt, gasTipCap ...int64) (*types.Transaction, error) {
	client := s.clientPool.GetClient()
	return BuildTransferERC20Tx(ctx, client, tokenAddress.ToETHAddress(), from.ToETHAddress(), to.ToETHAddress(), amount.ToBigInt(), gasTipCap...)
}

func (s *serviceImpl) GetAllowance(ctx context.Context, tokenAddress Address, ownerAddress, spenderAddress Address) (*dt.NumericInt, error) {
	client := s.clientPool.GetClient()
	allowanceFnSignature := []byte("allowance(address,address)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(allowanceFnSignature)
	methodID := hash.Sum(nil)[:4]

	paddedOwnerAddress := common.LeftPadBytes(ownerAddress.ToETHAddress().Bytes(), 32)
	paddedSpenderAddress := common.LeftPadBytes(spenderAddress.ToETHAddress().Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedOwnerAddress...)
	data = append(data, paddedSpenderAddress...)

	allowance, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   lo.ToPtr(tokenAddress.ToETHAddress()),
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	return dt.NumFromBigInt(big.NewInt(0).SetBytes(allowance)), nil
}

func (s *serviceImpl) BuildApproveToken(ctx context.Context, tokenAddress Address, ownerAddress, spenderAddress Address, amount *dt.NumericInt, gasTipCap ...int64) (*types.Transaction, error) {
	client := s.clientPool.GetClient()
	approveFnSignature := []byte("approve(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(approveFnSignature)
	methodID := hash.Sum(nil)[:4]

	paddedSpenderAddress := common.LeftPadBytes(spenderAddress.ToETHAddress().Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.ToBigInt().Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedSpenderAddress...)
	data = append(data, paddedAmount...)
	return BuildTx(ctx, client, ownerAddress.ToETHAddress(), tokenAddress.ToETHAddress(), big.NewInt(0), data, gasTipCap...)
}

func (s *serviceImpl) WaitTxAndGetReceipt(ctx context.Context, txHash TxHash, timeout ...time.Duration) (*types.Receipt, error) {
	timeoutDuration := time.Second * 30
	if len(timeout) > 0 {
		timeoutDuration = timeout[0]
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, timeoutCtx.Err()
		default:
			receipt, err := s.clientPool.TransactionReceipt(timeoutCtx, common.HexToHash(string(txHash)))
			if err != nil && !errors.Is(err, ethereum.NotFound) {
				return nil, err
			}

			if receipt != nil {
				if receipt.Status == types.ReceiptStatusSuccessful {
					return receipt, nil
				}

				return nil, ErrTxFailed
			}

			time.Sleep(time.Second)
		}
	}
}
