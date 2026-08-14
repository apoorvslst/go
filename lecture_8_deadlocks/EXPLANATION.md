# 🔄 Lecture 8: Handling Deadlocks in Database Transactions

## 🌟 High Level Overview (Ye Lecture Kya Hai?)
Is lecture mein hum seekhenge ki **Database Deadlocks** kya hote hain aur unko kaise avoid kiya jaata hai.
Jab do transactions ek hi waqt (concurrently) chal rahi hoti hain aur ek doosre ka infinite time ke liye wait karne lagti hain, tab deadlock aata hai. Is video mein humne dekha ki money transfer transaction mein ek potential deadlock tha, jise pehle humne test mein reproduce kiya, aur phir use completely fix kiya.

### ⚔️ Deadlock Kaise Aata Hai?
Socho do transactions ek hi time pe chal rahi hain:
- **Transaction 1 (Tx 1)**: Account 1 se Account 2 mein $10 transfer kar rahi hai.
- **Transaction 2 (Tx 2)**: Account 2 se Account 1 mein $10 transfer kar rahi hai (Reverse!).

Agar execution order aisa ho:
1. Tx 1: Account 1 ka balance subtract karti hai (Account 1 lock ho gaya).
2. Tx 2: Account 2 ka balance subtract karti hai (Account 2 lock ho gaya).
3. Tx 1: Ab Account 2 mein balance add karne jati hai -> **BLOCKED!** (Kyunki Tx 2 ne Account 2 pe already lock liya hua hai).
4. Tx 2: Ab Account 1 mein balance add karne jati hai -> **BLOCKED!** (Kyunki Tx 1 ne Account 1 pe pehle se lock liya hua hai).

Ab dono ek doosre ka wait kar rahe hain aur aage nahi badh sakte. **Boom! 💥 Deadlock!**

---

## 🛠️ Step-by-Step Implementation

### Step 1: Reproduce Deadlock in Test (`store_test.go`)
Deadlock fix karne se pehle usko code mein reproduce karna zaroori hai. Hum ek naya test function `TestTransferTxDeadlock` banayenge.

```go
func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n := 10 // 10 concurrent transactions chalayenge
	amount := int64(10)
	errs := make(chan error)

	for i := 0; i < n; i++ {
		// Default direction: Account 1 -> Account 2
		fromAccountID := account1.ID
		toAccountID := account2.ID

		// Half transactions reverse direction mein jayengi (Account 2 -> Account 1)
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

	// ... Verify errors and final balances ...
}
```

**📝 Code Explanation:**
- Hum `n=10` transactions background mein (`go func()`) chala rahe hain.
- `if i%2 == 1`: Is condition ki wajah se 5 transactions Account 1 -> Account 2 me paisa bhejengi, aur baaki 5 transactions Account 2 -> Account 1 me bhejengi.
- Jab hum ise run karenge toh purane code ke saath ye pakka **fail** hoga aur Postgres deadlock error dega!

---

### Step 2: Fix the Logic - Consistent Locking Order
Deadlock se bachne ka sabse best rule: **Humesha tables/rows ko ek consistent order mein lock karo.**
Yani, agar hum hamesha *choti ID wale account ko pehle update/lock karein*, toh deadlock kabhi nahi hoga! 

Code saaf rakhne ke liye pehle ek helper function banate hain `addMoney` jo 2 accounts ko ek sath update karega:

```go
// addMoney is a helper function to add money to two accounts
func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return // Naked return (named returns automatically return ho jayenge)
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return // Naked return
}
```

**📝 Code Explanation:**
- `addMoney` function bas do baar `AddAccountBalance` query ko call karta hai sequentially.
- Function signature mein humne returns ko variables ke naam de diye hain: `(account1 Account, account2 Account, err error)`.
- Isiliye Go syntax ke according, hum bina variables ke seedha `return` likh sakte hain (**Naked Return**). Go apne aap check karega ki in variables mein us point pe kya values hain aur wahi return kar dega.

---

### Step 3: Use the Helper Function with Ordered IDs (`store.go`)
Ab hum `TransferTx` function mein jahan accounts update ho rahe the, wahan check lagayenge ki kaunsa account ID chota hai.

```go
		// Update accounts balances
		// Hamesha SMALLER ID wale account ko pehle update karo taaki deadlocks na ho!
		
		if arg.FromAccountID < arg.ToAccountID {
			// Agar From account ID choti hai, toh usko pehle do
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			// Agar To account ID choti hai, toh usko pehle do
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}
```

**📝 Code Explanation:**
- `if arg.FromAccountID < arg.ToAccountID`: Agar From Account ki ID chhoti hai, toh `addMoney` ka first argument From Account hoga (subtract karne ke liye `-arg.Amount`), aur second To Account hoga.
- `else`: Agar To Account ki ID chhoti hai, toh `addMoney` mein pehle To Account pass hoga (add karne ke liye `arg.Amount`), phir From Account.
- **Kamaal ka nateeja:** Ab dono Tx 1 aur Tx 2 pehle chhoti ID wale account (e.g. Account 1) ko mangengi. Jo transaction Account 1 pehle jeetegi, woh pura process khatam karegi, aur tab tak doosri transaction patiently wait karegi. Deadlock solved! 🥳

Ab `TestTransferTxDeadlock` run karo, saare tests green aur fully pass ho jayenge!
