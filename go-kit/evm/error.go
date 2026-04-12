package evm

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
)

const ErrCodeJSONRPCServerErrorNodeUnhealthy = -32005

func isLogTooLargeError(err error) bool {
	if err == nil {
		return false
	}
	var e rpc.Error
	if (errors.As(err, &e) && e.ErrorCode() == ErrCodeJSONRPCServerErrorNodeUnhealthy && strings.Contains(err.Error(), "10000")) || // rate limit error infura
		strings.Contains(err.Error(), "range too large") { // rate limit cloudflare-eth.com
		return true
	}
	return false
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	var e rpc.Error
	if errors.As(err, &e) && e.ErrorCode() == ErrCodeJSONRPCServerErrorNodeUnhealthy && strings.Contains(err.Error(), "limit exceeded") ||
		strings.Contains(err.Error(), "429 Too Many Requests") {
		return true
	}
	return false
}
