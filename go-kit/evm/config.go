package evm

type Config struct {
	// List of RPC URL with comma separate
	RPCUrls         []string `json:"rpc_urls" mapstructure:"rpc_urls"`
	ManualBlockTime bool     `json:"manual_block_time" mapstructure:"manual_block_time"`
	DefaultProofs   int      `json:"num_of_proof" mapstructure:"num_of_proof"`
}
