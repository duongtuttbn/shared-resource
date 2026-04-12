package database

import (
	"fmt"
)

// PostgresqlConfig hold config for postgresql.
type PostgresqlConfig struct {
	Host     string `json:"host" mapstructure:"host" yaml:"host"`
	Database string `json:"database" mapstructure:"database" yaml:"database"`
	Port     int    `json:"port" mapstructure:"port" yaml:"port"`
	Username string `json:"username" mapstructure:"username" yaml:"username"`
	Password string `json:"password" mapstructure:"password" yaml:"password"`
	Options  string `json:"options" mapstrucuture:"options" yaml:"options"`
}

// String return Postgres connection string.
func (m PostgresqlConfig) String() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?%s", m.Username, m.Password, m.Host, m.Port, m.Database, m.Options)
}

func (m PostgresqlConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d %s", m.Host, m.Username, m.Password, m.Database, m.Port, m.Options,
	)
}
