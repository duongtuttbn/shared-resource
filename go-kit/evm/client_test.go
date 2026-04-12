package evm

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	client, err := NewClient("https://bsc-dataseed1.binance.org", "")
	if err != nil {
		t.Errorf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	block, err := client.BlockNumber(context.Background())
	if err != nil {
		t.Errorf("Failed to retrieve genesis block: %v", err)
	}
	if block <= 0 {
		t.Errorf("Invalid latest block number: have %v", block)
	}
	t.Logf("%v\n", block)
}

func TestManualBlockTime(t *testing.T) {
	cfg := Config{
		RPCUrls: []string{"https://rpc-mainnet.matic.network"},
	}
	clientPool, err := NewBasicClientPool(cfg)
	require.Nil(t, err)
	time := clientPool.manualBlockTime(8852480)
	log.Println(time)
}

func TestDoWithProofs(t *testing.T) {
	dummyErr := errors.New("dummy error")
	res, err := doWithProofs(2, func(index int) (int, error) {
		if index%2 == 0 {
			return index, nil
		}
		return 0, dummyErr
	}, func(a, b int) bool {
		return a == b
	})

	require.ErrorIs(t, err, ErrInconsistentProofs)
	require.Equal(t, 0, res)

	res, err = doWithProofs(2, func(_ int) (int, error) {
		return 0, ethereum.NotFound
	}, func(a, b int) bool {
		return a == b
	})

	require.NotErrorIs(t, err, ErrInconsistentProofs)
	require.ErrorIs(t, err, ethereum.NotFound)
	require.Equal(t, 0, res)

	res, err = doWithProofs(2, func(_ int) (int, error) {
		return 1, nil
	}, func(a, b int) bool {
		return a == b
	})

	require.NotErrorIs(t, err, ErrInconsistentProofs)
	require.NoError(t, err)
	require.Equal(t, 1, res)
}
