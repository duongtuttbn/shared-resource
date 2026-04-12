package evm

import (
	"context"
	"math/big"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	*ethclient.Client
	lastErr     error
	availableAt time.Time
	mu          sync.Mutex
	rpcClient   *rpc.Client
	endpoint    string
}

func NewClient(endpoint, proxyURL string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse endpoint to url")
	}
	switch u.Scheme {
	case "http", "https":
		return NewHTTPClient(endpoint, proxyURL)
	default:
		return NewUniversalClient(endpoint)
	}
}

func NewUniversalClient(endpoint string) (*Client, error) {
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "unable to dial endpoint with eth client")
	}
	return &Client{
		Client:      client,
		lastErr:     nil,
		availableAt: time.Now(),
		endpoint:    endpoint,
	}, nil
}

func NewHTTPClient(endpoint string, proxyURL string) (*Client, error) {
	httpClient := &http.Client{}
	if proxyURL != "" {
		u, err := url.Parse(proxyURL)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse proxyURL to url")
		}
		httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
	}

	rpcClient, err := rpc.DialOptions(context.Background(), endpoint, rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, errors.Wrap(err, "unable to dial endpoint with rpc")
	}

	return &Client{
		Client:      ethclient.NewClient(rpcClient),
		lastErr:     nil,
		availableAt: time.Now(),
		rpcClient:   rpcClient,
		endpoint:    endpoint,
	}, nil
}

func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lastErr == nil || time.Now().After(c.availableAt) {
		c.lastErr = nil
		return true
	}
	return false
}

func (c *Client) MarkError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastErr = err
	// client will be available after 1 minute
	c.availableAt = time.Now().Add(time.Minute)
}

func (c *Client) GetRPCClient() *rpc.Client {
	return c.rpcClient
}

func (c *Client) EndpointURL() string {
	return c.endpoint
}

func (c *Client) GetERC20Balance(ctx context.Context, tokenAddress, address common.Address) (*big.Int, error) {
	balanceOfFnSignature := []byte("balanceOf(address)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(balanceOfFnSignature)
	methodID := hash.Sum(nil)[:4]

	paddedAddress := common.LeftPadBytes(address.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)

	balance, err := c.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}
	return big.NewInt(0).SetBytes(balance), nil
}
