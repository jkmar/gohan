package goplugin

import (
	"context"

	gohan_db "github.com/cloudwan/gohan/db"
	"github.com/cloudwan/gohan/db/transaction"
	"github.com/cloudwan/gohan/extension/goext"
)

// Database in an implementation of IDatabase
type Database struct {
	db gohan_db.DB
}

// NewDatabase creates new database implementation
func NewDatabase(env *Environment) goext.IDatabase {
	return &Database{db: env.db}
}

// Begin starts a new transaction
func (db *Database) Begin() (goext.ITransaction, error) {
	t, _ := db.db.Begin()
	return &Transaction{t}, nil
}

// BeginTx starts a new transaction with options
func (db *Database) BeginTx(ctx goext.Context, options *goext.TxOptions) (goext.ITransaction, error) {
	opts := transaction.TxOptions{IsolationLevel: transaction.Type(options.IsolationLevel)}
	t, _ := db.db.BeginTx(context.Background(), &opts)
	return &Transaction{t}, nil
}