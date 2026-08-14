package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTransferTx concurrent transfer transaction ko test karta hai
func TestTransferTx(t *testing.T) {
	// Naya store object banate hain hamare testDB se (jo main_test.go me declare kiya tha)
	store := NewStore(testDB)

	// Dono accounts (Sender aur Receiver) banate hain random functions se
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// Concurrency handle karne ke liye (bahut saari requests ek sath), hum channels use karenge
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	// 'n' times go routines chalayenge (background threads ki tarah samajho)
	for i := 0; i < n; i++ {
		go func() {
			// Har ek routine me transaction fire hoga
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			
			// Main routine ko error aur result waapis bhejne ke liye channels use karte hain
			errs <- err
			results <- result
		}()
	}

	// Ab main routine me wait karenge saare 'n' results aane ka
	for i := 0; i < n; i++ {
		err := <-errs // channel se error liya
		require.NoError(t, err)

		result := <-results // channel se result liya
		require.NotEmpty(t, result)

		// Check the transfer record
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		// Confirm from DB if transfer exists
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// Check FromEntry (Jisme se paisa nikla hai)
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		// Check ToEntry (Jisme paisa aaya hai)
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// TODO: balance update ki checks next lecture me karenge
	}
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n := 10
	amount := int64(10)
	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// Check final updated balances
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}
