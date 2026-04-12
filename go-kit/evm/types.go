package evm

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

var (
	ErrTxFailed            = errors.New("transaction failed")
	ErrInsufficientBalance = errors.New("insufficient balance")
	NativeTokenAddress     = Address("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")
)

type ChainID int

func (c ChainID) String() string {
	return strconv.Itoa(int(c))
}

func (c ChainID) BigInt() *big.Int {
	return big.NewInt(int64(c))
}

type Address string

func (a Address) ToLower() Address {
	return Address(strings.ToLower(string(a)))
}

func (a Address) Equals(b Address) bool {
	return a.String() == b.String()
}

func (a Address) String() string {
	return string(a.toChecksum())
}

func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a Address) Value() (driver.Value, error) {
	return a.String(), nil
}

func (a Address) ToETHAddress() common.Address {
	return common.HexToAddress(string(a))
}

func (a Address) IsHexAddress() bool {
	return common.IsHexAddress(string(a))
}

func (a Address) toChecksum() Address {
	if !a.IsHexAddress() {
		return a
	}
	return Address(common.HexToAddress(string(a)).Hex())
}

type (
	UserOpHash string
	TxHash     string
)

type SignableWallet interface {
	Address() common.Address
	SignTx(ctx context.Context, tx *types.Transaction) (*types.Transaction, error)
}

type TokenInfo struct {
	Address  Address `json:"address"`
	Decimals int     `json:"decimals"`
}
