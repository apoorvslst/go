package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/apoor/simple_bank/util"
)

// createRandomAccount ek helper function hai jo database me ek naya account banata hai
// Isko hum dusre tests (Get, Update, Delete) ke andar use karte hain
func createRandomAccount(t *testing.T) Account {
	// util package use karke random details banate hain taaki har test ka data alag ho
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	// Actual CreateAccount function call karte hain
	account, err := testQueries.CreateAccount(context.Background(), arg)
	
	// require.NoError test ko fail kar dega agar koi error aati hai
	require.NoError(t, err)
	// check karte hain ki account empty nahi hona chahiye
	require.NotEmpty(t, account)

	// ensure karte hain ki jo details humne bheji thi, wahi save hui hain
	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	// ID aur CreatedAt database apne aap lagata hai, toh wo 0 ya empty nahi hone chahiye
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

// TestCreateAccount sirf Create operation ko test karne ke liye hai
func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

// TestGetAccount ek account ko retrieve karne ko test karta hai
func TestGetAccount(t *testing.T) {
	// Pehle ek account create karo (humein pata hona chahiye kya nikaalna hai)
	account1 := createRandomAccount(t)
	
	// Ab uss account1 ki ID use karke usko database se nikaalo
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account2)

	// Phir compare karo ki fetched account ki details create kiye account ke barabar ho
	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	// Timestamps ekdum same nahi hote, so hum check karte hain ki unme difference 1 second se kam ho
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
}

// TestUpdateAccount check karta hai account updates
func TestUpdateAccount(t *testing.T) {
	// Pehle account banaya
	account1 := createRandomAccount(t)

	// Naya balance random choose kiya
	arg := UpdateAccountParams{
		ID:      account1.ID,
		Balance: util.RandomMoney(),
	}

	// Update query chalayi
	account2, err := testQueries.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account2)

	// Check kiya ki id, owner, currency wahi ho, but BALANCE update ho gaya ho
	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, arg.Balance, account2.Balance) 
	require.Equal(t, account1.Currency, account2.Currency)
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
}

// TestDeleteAccount check karta hai ki account successfully delete hota hai
func TestDeleteAccount(t *testing.T) {
	// 1. Ek account banaya
	account1 := createRandomAccount(t)
	
	// 2. Us account ko delete kiya
	err := testQueries.DeleteAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	// 3. Ab try kiya usko wapis nikaalne ka (should fail)
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	
	// Ensure karte hain ki error aa gayi hai
	require.Error(t, err)
	// Error exactly "sql.ErrNoRows" honi chahiye kyunki ab us id ka koi data nahi hai
	require.EqualError(t, err, sql.ErrNoRows.Error())
	// Aur return hua account object completely empty hona chahiye
	require.Empty(t, account2)
}

// TestListAccounts check karta hai ki list fetch sahi ho rahi hai pagination ke sath
func TestListAccounts(t *testing.T) {
	// 10 random accounts banaye 
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	// Limit 5 aur Offset 5 ke sath parameters set kiye
	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	// ListAccounts chalaya
	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	// List ki length 5 honi chahiye
	require.Len(t, accounts, 5)

	// Aur har ek account properly exist karna chahiye
	for _, account := range accounts {
		require.NotEmpty(t, account)
	}
}
