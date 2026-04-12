package redis

type KVConfig struct {
	Prefix string
}

type KVOpt func(c *KVConfig)

func WithPrefix(prefix string) KVOpt {
	return func(c *KVConfig) {
		c.Prefix = prefix
	}
}
