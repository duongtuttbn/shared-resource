package postgresfx

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"tla-backend/pkg/go-kit/db"
)

// NewModule behaves like gormx.NewModule with PostgreSQL dialect registered.
func NewModule() fx.Option {
	return fx.Module(
		"postgresfx",
		fx.Provide(
			NewDialector,
			NewDatabase,
			db.NewTransactionManager,
		),
	)
}

// NewDatabase create a *db.Database with specified config.
func NewDatabase(p DatabaseParams) (*db.DB, error) {
	gormConf := lo.CoalesceOrEmpty(p.Default, &gorm.Config{})
	if p.Config.GormTranslateError {
		gormConf.TranslateError = true
	}
	if p.Config.GormDisableDefaultTransaction {
		gormConf.SkipDefaultTransaction = true
	}
	gormDB, err := gorm.Open(p.Dialector, gormConf)
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	if p.Config.GormConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(p.Config.GormConnMaxLifetime)
	}
	if p.Config.GormConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(p.Config.GormConnMaxIdleTime)
	}
	if p.Config.GormMaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(p.Config.GormMaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(100)
	}
	if p.Config.GormMaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(p.Config.GormMaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10)
	}
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			d, err := gormDB.DB()
			if err != nil {
				return err
			}
			err = d.Close()
			return err
		},
	})
	return db.NewDB(gormDB), nil
}

// NewDialector create with PostgreSQL gorm.Dialector.
func NewDialector(c GormConfig) (gorm.Dialector, error) {
	return postgres.Open(c.Postgres.DSN()), nil
}

// IsPgError check whether the error is any of *pgconn.ConnectError or *pgconn.PgError.
func IsPgError(err error) bool {
	var pqerr *pgconn.ConnectError
	if ok := errors.As(err, &pqerr); ok {
		return true
	}
	var connErr *pgconn.PgError
	if ok := errors.As(err, &connErr); ok {
		return true
	}
	return false
}
