package postgres

import (
	"context"
	"database/sql"
)

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	ctxTx := withTx(ctx, tx)

	if err := fn(ctxTx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
