package evm

import (
	"context"
	errors2 "errors"
	"fmt"
	"math/big"
	"sync"
	"time"
	"github.com/duongtuttbn/shared-resource/go-kit/log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

var ErrInconsistentProofs = errors.New("inconsistent proofs")

type ClientPool struct {
	clients       []*Client
	counter       int
	mu            sync.Mutex
	config        Config
	logger        log.Logger
	defaultProofs int
}

func NewBasicClientPool(cfg Config) (*ClientPool, error) {
	return NewBasicClientPoolWithLogger(cfg, log.Root())
}

func NewBasicClientPoolWithLogger(cfg Config, logger log.Logger) (*ClientPool, error) {
	clients := make([]*Client, len(cfg.RPCUrls))
	for i, rpcURL := range cfg.RPCUrls {
		var err error
		clients[i], err = NewClient(rpcURL, "")
		if err != nil {
			return nil, errors.Wrap(err, "unable to init new client")
		}
	}

	return NewClientPool(clients, cfg, logger), nil
}

func NewClientPool(clients []*Client, config Config, logger log.Logger) *ClientPool {
	defaultProofs := 1
	if config.DefaultProofs > 0 {
		defaultProofs = config.DefaultProofs
	}
	return &ClientPool{
		clients:       clients,
		config:        config,
		logger:        logger,
		defaultProofs: defaultProofs,
	}
}

// GetClient block until a client is available
func (p *ClientPool) GetClient() *Client {
	p.mu.Lock()
	defer p.mu.Unlock()
	for {
		for i := 0; i < len(p.clients); i++ {
			client := p.clients[p.counter]
			p.counter = (p.counter + 1) % len(p.clients)
			if client.IsAvailable() {
				return client
			}
		}
		p.logger.Info("all clients are down, sleep for 1 minute")
		time.Sleep(time.Minute)
	}
}

func (p *ClientPool) GetClients(numClients int) ([]*Client, error) {
	if numClients > len(p.clients) {
		return nil, fmt.Errorf("numClients must be less than client pool size: %d", len(p.clients))
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for {
		availableClients := p.countAvailableClients()
		if availableClients < numClients {
			p.logger.Info(fmt.Sprintf("Request %d clients but only %d available. Sleep for 1 minute", numClients, availableClients))
			time.Sleep(time.Minute)
			continue
		}
		break
	}

	clients := make([]*Client, 0, numClients)
	for len(clients) < numClients {
		client := p.clients[p.counter]
		p.counter = (p.counter + 1) % len(p.clients)
		if client.IsAvailable() {
			clients = append(clients, client)
		}
	}
	return clients, nil
}

func (p *ClientPool) RunOp(ctx context.Context, op func(client *Client) error) {
	for ctx.Err() == nil {
		client := p.GetClient()
		err := op(client)
		if err == nil {
			return
		}
	}
}

func (p *ClientPool) GetLatestBlock() uint64 {
	for {
		client := p.GetClient()
		maxBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			client.MarkError(err)
			p.logger.Errorf("get max block error: %v", err)
			continue
		}
		return maxBlock
	}
}

func (p *ClientPool) GetLogs(contractAddresses []common.Address, filterTopics [][]common.Hash, fromBlock, toBlock uint64, numProof ...int) ([]types.Log, error) {
	proofs := p.defaultProofs
	if len(numProof) > 0 {
		proofs = numProof[0]
	}
	availableClients, err := p.GetClients(proofs)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get clients")
	}

	result, err := doWithProofs(
		proofs,
		func(index int) ([]types.Log, error) {
			return p.getLogs(contractAddresses, filterTopics, fromBlock, toBlock, availableClients[index])
		},
		p.compareListsLogs,
		func(i int, result []types.Log) {
			availableClients[i].MarkError(ErrInconsistentProofs)
			p.logger.Errorf("[Consistency error] Block range: [%d, %d]: Client %s returned %d logs", fromBlock, toBlock, availableClients[i].endpoint, len(result))
		},
	)
	if err != nil {
		if errors.Is(err, ErrInconsistentProofs) {
			p.logger.Info("Try to get logs again by other clients...")
			return p.GetLogs(contractAddresses, filterTopics, fromBlock, toBlock, proofs)
		}
		return nil, err
	}
	return result, nil
}

func (p *ClientPool) BlockTime(blockNumber uint64) uint64 {
	if p.config.ManualBlockTime {
		return p.manualBlockTime(blockNumber)
	}
	return p.rpcBlockTime(blockNumber)
}

func (p *ClientPool) TransactionReceipt(ctx context.Context, txHash common.Hash, numProof ...int) (*types.Receipt, error) {
	proofs := p.defaultProofs
	if len(numProof) > 0 {
		proofs = numProof[0]
	}

	clients, err := p.GetClients(proofs)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get clients")
	}

	result, err := doWithProofs(
		proofs,
		func(index int) (*types.Receipt, error) {
			return clients[index].TransactionReceipt(ctx, txHash)
		},
		p.compareReceipt,
		func(index int, _ *types.Receipt) {
			clients[index].MarkError(ErrInconsistentProofs)
		},
	)
	if err != nil {
		if errors.Is(err, ErrInconsistentProofs) {
			return p.TransactionReceipt(ctx, txHash, proofs)
		}
		return nil, err
	}

	return result, nil
}

func (p *ClientPool) compareListsLogs(logs1 []types.Log, logs2 []types.Log) (equal bool) {
	if len(logs1) != len(logs2) {
		p.logger.Debug("logs1 (%d items) is not equal to logs2 (%d items)", len(logs1), len(logs2))
		return false
	}

	for i := 0; i < len(logs1); i++ {
		if !p.compareLogs(logs1[i], logs2[i]) {
			p.logger.Debug("logs1[%d] is not equal to logs2[%d]. \n\tlog1: %v\n\tlog2: %v", i, i, logs1[i], logs2[i])
			return false
		}
	}
	return true
}

func (p *ClientPool) compareLogs(log1 types.Log, log2 types.Log) (equal bool) {
	return log1.TxHash == log2.TxHash &&
		log1.Index == log2.Index
}

func (p *ClientPool) getLogs(contractAddresses []common.Address, filterTopics [][]common.Hash, fromBlock, toBlock uint64, specificClient ...*Client) ([]types.Log, error) {
	if fromBlock > toBlock {
		return []types.Log{}, nil
	}
	filter := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Topics:    filterTopics,
		Addresses: contractAddresses,
	}

	var client *Client
	if len(specificClient) > 0 && specificClient[0] != nil {
		client = specificClient[0]
	} else {
		client = p.GetClient()
	}
	logs, err := client.FilterLogs(context.Background(), filter)
	if err != nil {
		p.logger.Errorf("Fetch logs [%d to %d] on endpoint %v error: %v", fromBlock, toBlock, client.endpoint, err)
		if isLogTooLargeError(err) && toBlock > fromBlock {
			midBlockNumber := fromBlock + (toBlock-fromBlock)/2
			log1, err1 := p.getLogs(contractAddresses, filterTopics, fromBlock, midBlockNumber)
			if err1 != nil {
				return nil, err1
			}
			log2, err2 := p.getLogs(contractAddresses, filterTopics, midBlockNumber+1, toBlock)
			if err2 != nil {
				return nil, err2
			}
			log1 = append(log1, log2...)
			return log1, nil
		}
		client.MarkError(err)
		if isRateLimitError(err) {
			return p.getLogs(contractAddresses, filterTopics, fromBlock, toBlock)
		}

		return nil, errors.Wrap(err, "client endpoint: "+client.endpoint)
	}

	return logs, nil
}

func (p *ClientPool) rpcBlockTime(blockNumber uint64) uint64 {
	for {
		ethClient := p.GetClient()
		block, err := ethClient.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
		if err != nil {
			p.logger.Info(fmt.Sprintf("error requesting blocktime from node, backing off. BlockNumber: %v Endpoint: %v, Err: %v,", blockNumber, ethClient.endpoint, err))
			ethClient.MarkError(err)
			continue
		}
		return block.Time()
	}
}

type GetBlockTimeResponse struct {
	Result struct {
		Timestamp string `json:"timestamp"`
	} `json:"result"`
}

func (p *ClientPool) manualBlockTime(blockNumber uint64) uint64 {
	for {
		ethClient := p.GetClient()
		url := ethClient.endpoint
		body := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "eth_getBlockByNumber",
			"params": []interface{}{
				DecimalToHex(int64(blockNumber)),
				true,
			},
			"id": 0,
		}
		client := resty.New()
		res, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			SetResult(GetBlockTimeResponse{}).
			Post(url)
		if err != nil {
			p.logger.Info(fmt.Sprintf("error manual requesting blocktime from node, backing off. BlockNumber: %v Endpoint: %v, Err: %v,", blockNumber, ethClient.endpoint, err))
			ethClient.MarkError(err)
			continue
		}
		if res.IsError() {
			p.logger.Info(fmt.Sprintf("error manual requesting blocktime from node, status code error. BlockNumber: %v Endpoint: %v, Err: %v,", blockNumber, ethClient.endpoint, string(res.Body())))
			ethClient.MarkError(err)
			continue
		}
		data := res.Result().(*GetBlockTimeResponse)
		result, err := HexToInt(data.Result.Timestamp)
		if err != nil {
			p.logger.Info(fmt.Sprintf("error manual requesting blocktime from node, hex to int. BlockNumber: %v Endpoint: %v, Err: %v,", blockNumber, ethClient.endpoint, err))
			ethClient.MarkError(err)
			continue
		}
		return uint64(result)
	}
}

func (p *ClientPool) compareReceipt(a, b *types.Receipt) bool {
	if a == nil || b == nil {
		return false
	}

	if a.TxHash != b.TxHash {
		return false
	}

	if a.BlockNumber.Cmp(b.BlockNumber) != 0 {
		return false
	}

	if a.TransactionIndex != b.TransactionIndex {
		return false
	}

	if a.Status != b.Status {
		return false
	}

	return p.compareListsLogs(convertSlices(a.Logs, lo.FromPtr), convertSlices(b.Logs, lo.FromPtr))
}

func (p *ClientPool) countAvailableClients() int {
	counter := 0
	for _, client := range p.clients {
		if client.IsAvailable() {
			counter++
		}
	}
	return counter
}

func convertSlices[T any, R any](collection []T, iteratee func(item T) R) []R {
	result := make([]R, len(collection))

	for i := range collection {
		result[i] = iteratee(collection[i])
	}

	return result
}

func doWithProofs[T any](proofs int, fn func(index int) (T, error), compare func(a, b T) bool, onInconsistent ...func(index int, result T)) (T, error) {
	if proofs <= 1 {
		return fn(0)
	}

	var wg sync.WaitGroup
	wg.Add(proofs)
	results := make([]T, proofs)
	errorList := make([]error, proofs)
	for i := 0; i < proofs; i++ {
		go func(index int) {
			defer wg.Done()
			result, subErr := fn(index)
			if subErr != nil {
				errorList[index] = subErr
			}
			results[index] = result
		}(i)
	}
	wg.Wait()
	var empty T
	err := errorList[0]
	for i := 1; i < proofs; i++ {
		if !errors.Is(err, errorList[i]) {
			return empty, errors2.Join(ErrInconsistentProofs, errors2.Join(errorList...))
		}
	}
	if err != nil {
		return empty, err
	}

	consistent := false
	for i := 0; i < proofs; i++ {
		if i < proofs-1 {
			consistent = compare(results[i], results[i+1])
			if !consistent {
				break
			}
		}
	}

	if consistent {
		return results[0], nil
	}

	if len(onInconsistent) > 0 {
		for i := 0; i < proofs; i++ {
			onInconsistent[0](i, results[i])
		}
	}

	return empty, ErrInconsistentProofs
}
