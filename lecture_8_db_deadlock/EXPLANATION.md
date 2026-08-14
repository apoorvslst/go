# 🏦 Lecture 8: Handling Database Deadlocks in Golang

---

## 🦅 High-Level Overview (Bird's Eye View)

Imagine karo 2 log ek doosre ko exact same time par paise bhej rahe hain:
- **Transaction 1:** Apoorv (Account 1) ne Rahul (Account 2) ko ₹10 bheje.
- **Transaction 2:** Rahul (Account 2) ne Apoorv (Account 1) ko ₹10 bheje.

**Problem (The Deadlock):**
Transaction 1 pehle Apoorv ka account update karne ke liye "lock" lagata hai, phir Rahul ke account pe jayega.
Transaction 2 pehle Rahul ka account "lock" karta hai, phir Apoorv ke account pe jayega.

Agar dono transactions ek saath parallel chalein:
1. Tx1 ne Apoorv ko lock kar liya.
2. Tx2 ne Rahul ko lock kar liya.
3. Ab Tx1 aage badhne ke liye Rahul ko lock karna chahta hai, jo ki Tx2 ke paas hai -> Wait karega.
4. Waise hi Tx2 aage badhne ke liye Apoorv ko lock karna chahta hai, jo ki Tx1 ke paas hai -> Wait karega.

Dono ek doosre ka wait karte reh jayenge aur system atak jayega. Isko bolte hain **Deadlock**!

**Solution:**
Humesha "Locks" ek **consistent order** (ek taye kram) mein lo. 
Example: Humesha chote Account ID ko pehle update karo. Agar Apoorv ka ID `1` hai aur Rahul ka `2`, toh chahe paisa koi bhi bheje (A->R ya R->A), pehle ID `1` lock hoga phir ID `2`. Isse deadlock kabhi hoga hi nahi!

**Code mein kya kiya?**
1. **Understand:** Psql console aur TablePlus ki madad se manually queries run karke deadlock banakar dekha ki ye hota kaise hai.
2. **Replicate:** `store_test.go` mein ek naya test `TestTransferTxDeadlock` banaya jisme 10 parallel transactions chalayin (5 normal, 5 reverse) taaki deadlock fail hone lag jaye.
3. **Fix:** `store.go` ke `TransferTx` function mein `if-else` lagaya taaki humesha chota Account ID pehle update/lock ho.
4. **Refactor:** Code ki duplication hatane ke liye `addMoney()` helper function banaya.

---

## 🗺️ Step-by-Step: Kya Pehle Banao, Kya Baad Mein?

Agar tum ye lecture khud se follow karna chaho toh ye order hai:

### Step 1: Psql Console mein Deadlock Samajhna
2 alag-alag terminal tabs mein psql console open kiya aur manual `BEGIN` se transactions run ki. Reverse flow mein jab update commands chali, tab dekha ki query block/hang ho jaati hai. TablePlus mein ja kar locks ki list dekhi.

### Step 2: `store_test.go` mein Deadlock Test Banao
📁 **Edit karo:** `db/sqlc/store_test.go`
- Naya test function banaya: `TestTransferTxDeadlock`
- 10 goroutines chalaye jisme logic aisa rakha ki 5 aage ki direction mein paise bheje aur 5 ulti direction mein.
- Test abhi fail hoga kyunki "deadlock detected" error aayegi.

### Step 3: `store.go` mein Order Fix Karo (The Actual Fix)
📁 **Edit karo:** `db/sqlc/store.go`
- `TransferTx` function mein dono accounts ke update wale logic ko check karke choti ID ke hisaab se order kiya:
```go
if arg.FromAccountID < arg.ToAccountID {
    // update fromAccount pehle, fir toAccount
} else {
    // update toAccount pehle, fir fromAccount
}
```

### Step 4: Code ko Refactor Karo (`addMoney`)
📁 **Edit karo:** `db/sqlc/store.go`
- Ek naya helper function `addMoney` banaya jo baar-baar dono account balance update ka duplicate code hata de.
- Go (Golang) ka "named return variables" feature ka fayda uthaya.

### Step 5: Final Test Run Karo
```bash
go test -v -cover ./...
```
Sab green aana chahiye! Deadlock khatam! ✅

---

## 🛠 Code Explanation 1: Deadlock Replicating Test (`store_test.go`)

Humne pichle lecture wale `TestTransferTx` ko copy karke naya `TestTransferTxDeadlock` banaya. Isme main twist yeh condition hai:

```go
fromAccountID := account1.ID
toAccountID := account2.ID

// Agar loop counter 'i' odd (1, 3, 5..) hai, toh reverse the direction!
if i%2 == 1 {
	fromAccountID = account2.ID
	toAccountID = account1.ID
}
```

Kyunki 5 transactions normal hain aur 5 reverse, last mein accounts ka initial balance aur updated balance exact barabar aana chahiye (net exchange 0 hua).

```go
// Koi final error nahi honi chahiye
require.Equal(t, account1.Balance, updatedAccount1.Balance)
require.Equal(t, account2.Balance, updatedAccount2.Balance)
```
*Isko fix lagane se pehle run karoge toh fail hoga.*

---

## 🏗 Code Explanation 2: Preventing Deadlocks (`store.go`)

Sabse best tareeka deadlock rokne ka hai "Consistent Order". Hum ID ko compare karke choti ID wale account ko hamesha pehle update (aur therefore lock) karenge.

```go
if arg.FromAccountID < arg.ToAccountID {
	// FromAccount ki ID choti hai, toh pehle ise update karo
	result.FromAccount, result.ToAccount, err = addMoney(
		ctx, q, 
		arg.FromAccountID, -arg.Amount, // Paisa nikla
		arg.ToAccountID, arg.Amount,    // Paisa aaya
	)
} else {
	// ToAccount ki ID choti hai, toh pehle ToAccount ko update karo
	result.ToAccount, result.FromAccount, err = addMoney(
		ctx, q, 
		arg.ToAccountID, arg.Amount,    // Paisa aaya
		arg.FromAccountID, -arg.Amount, // Paisa nikla
	)
}
```

---

## 💸 Code Explanation 3: Refactoring with `addMoney` function

Upar wale fix ke baad humein account update ka code baar baar likhna padh raha tha. Toh ek helper function bana diya jisme Go ki pyari si "naked return" syntax use ki:

```go
func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) { // Yahan return variables ke naam pehle hi de diye

	// Pehle Account 1 ka balance update
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return // "Naked return": account1, account2, aur err automatically return ho jayenge (err ki wajah se ye yahi se vapis chala jayega)
	}

	// Phir Account 2 ka balance update
	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	
	return // Sab badiya, account1, account2 aur err automatically return!
}
```

Ye "naked return" Go language ki bohot concise feature hai. Jab hum return variables (`account1, account2, err`) function signature mein hi bata dete hain, toh `return` akele likhne par wo variables jis bhi state mein hote hain, return kar diye jate hain.
