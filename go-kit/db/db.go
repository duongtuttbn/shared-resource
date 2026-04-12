package db

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

var ErrNotFound = gorm.ErrRecordNotFound

type DB struct {
	db *gorm.DB
}

func NewDB(db *gorm.DB) *DB {
	return &DB{
		db: db,
	}
}

func (b *DB) DB(ctx context.Context) *gorm.DB {
	if ctx, ok := ctx.(*TransactedContext); ok {
		return ctx.tx
	}
	return b.db.WithContext(ctx)
}

func (b *DB) Table(ctx context.Context, tableName string, args ...interface{}) *gorm.DB {
	return b.DB(ctx).Table(tableName, args...)
}

type TransactedContext struct {
	context.Context
	tx *gorm.DB
}

func (t *TransactedContext) Commit() error {
	return t.tx.Commit().Error
}

func (t *TransactedContext) Rollback() error {
	return t.tx.Rollback().Error
}

func (t *TransactedContext) SavePoint(name string) error {
	return t.tx.SavePoint(name).Error
}

func (t *TransactedContext) RollbackTo(name string) error {
	return t.tx.RollbackTo(name).Error
}

func (t *TransactedContext) Transaction(fc func(tx *TransactedContext) error) error {
	return t.tx.Transaction(func(tx *gorm.DB) error {
		return fc(&TransactedContext{
			Context: t.Context,
			tx:      tx,
		})
	})
}

// TransactionManager provide away to start a Transaction context without using Database.
type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *DB) *TransactionManager {
	return &TransactionManager{db: db.db}
}

func (t *TransactionManager) Begin(ctx context.Context, opts ...*sql.TxOptions) (*TransactedContext, error) {
	tx := t.db.WithContext(ctx).Begin(opts...)
	return &TransactedContext{
		Context: ctx,
		tx:      tx,
	}, tx.Error
}

func (t *TransactionManager) Transaction(ctx context.Context, fc func(ctx context.Context) error, opts ...*sql.TxOptions) error {
	if ctx, ok := ctx.(*TransactedContext); ok {
		return ctx.tx.Transaction(func(tx *gorm.DB) error {
			return fc(&TransactedContext{
				Context: ctx.Context,
				tx:      tx,
			})
		})
	}

	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fc(&TransactedContext{
			Context: ctx,
			tx:      tx,
		})
	}, opts...)
}

func Paging(db *gorm.DB, page, limit uint) *gorm.DB {
	if limit > 0 {
		db.Limit(int(limit))
		if page > 0 {
			db.Offset(int((page - 1) * limit))
		}
	}

	return db
}
