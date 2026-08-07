package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store struct saari queries (individual & transaction) run karne ka method dega
type Store struct {
	*Queries
	db *sql.DB
}

// NewStore ek naya Store banata hai using sql.DB connection
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db), // Ye default sqlc generate kiya function hai
	}
}

// execTx ek transaction start karta hai aur callback function us transaction ke andar chalata hai
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	// Pehle transaction shuru karo
	tx, err := store.db.BeginTx(ctx, nil) // nil for default isolation level
	if err != nil {
		return err
	}

	// Us transaction ka Queries object banao
	q := New(tx)
	
	// Ab callback (jo data change karega) ko run karo
	err = fn(q)
	if err != nil {
		// Agar koi error aaya to Rollback (undo) karo!
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	// Sab successful raha to Commit (permanently save) kardo
	return tx.Commit()
}

// TransferTxParams wo inputs hain jo hume money transfer ke liye chahiye
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult wo output hai jo transaction successful hone pe aayega
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx money transfer karta hai (ek hi transaction mein record, 2 entries, aur dono ke balance update)
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		
		// 1. Transfer record banaya
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// 2. FromAccount ki Entry banayi (paisa nikalne ki)
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// 3. ToAccount ki Entry banayi (paisa aane ki)
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// TODO: balance updating will be implemented in the next lecture (Deadlocks!)

		return nil
	})

	return result, err
}
