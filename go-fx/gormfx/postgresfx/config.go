package postgresfx

import (
	"fmt"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"time"
	"github.com/duongtuttbn/shared-resource/go-kit/log"
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

type GormConfig struct {
	Postgres PostgresqlConfig `json:"postgres" mapstructure:"postgres"`

	GormTranslateError            bool `json:"gorm_translate_error" mapstructure:"gorm_translate_error"`
	GormDisableDefaultTransaction bool `json:"gorm_disable_default_transaction" mapstructure:"gorm_disable_default_transaction"`

	GormConnMaxLifetime time.Duration `json:"gorm_conn_max_lifetime" mapstructure:"gorm_conn_max_lifetime"`
	GormConnMaxIdleTime time.Duration `json:"gorm_conn_max_idle_time" mapstructure:"gorm_conn_max_idle_time"`

	GormMaxOpenConns int `json:"gorm_max_open_conns" mapstructure:"gorm_max_open_conns"`
	GormMaxIdleConns int `json:"gorm_max_idle_conns" mapstructure:"gorm_max_idle_conns"`
}

type DatabaseParams struct {
	fx.In
	Config    GormConfig
	Logger    log.Logger   `optional:"true"`
	Default   *gorm.Config `optional:"true"`
	Dialector gorm.Dialector
	Lifecycle fx.Lifecycle
}
