package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

var TxKey = txKey{}

type Manager struct {
	db *sqlx.DB
}

func NewManager(db *sqlx.DB) *Manager {
	return &Manager{db: db}
}

func (m *Manager) Do(ctx context.Context, fn func(context.Context) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	txCtx := context.WithValue(ctx, TxKey, tx)
	if err = fn(txCtx); err != nil {
		return err
	}
	return tx.Commit()
}

func TxFromContext(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(TxKey).(*sqlx.Tx)
	return tx, ok
}
