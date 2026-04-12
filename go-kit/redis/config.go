package redis

type Config struct {
	Address  string `json:"address" mapstructure:"address"`
	Password string `json:"password" mapstructure:"password"`
	Database int    `json:"database" mapstructure:"database"`
}
