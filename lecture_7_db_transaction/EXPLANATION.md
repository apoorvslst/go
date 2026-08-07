# 🏦 Lecture 7: Database Transactions in Golang

---

## 🦅 High-Level Overview (Bird's Eye View)

Imagine karo tum **PhonePe/GPay** se kisi ko ₹100 bhej rahe ho.

Ye ek "simple" action lagta hai, lekin andar se **5 cheezein** honi chahiye:
1. Ek **transfer record** bane (ki "Apoorv ne Rahul ko ₹100 bheje")
2. Apoorv ki **entry** bane (ki "Apoorv se ₹100 nikle")
3. Rahul ki **entry** bane (ki "Rahul ko ₹100 aaye")
4. Apoorv ka **balance** ₹100 ghatao
5. Rahul ka **balance** ₹100 badhao

**Problem:** Soch lo Step 3 ke baad server crash ho gaya. Apoorv ke ₹100 kat gaye, Rahul ko mile nahi — paisa hawa mein gayab!

**Solution: Transaction** → "Ya toh ye paancho steps saari hongi, ya ek bhi nahi hogi."
- Fail hua? → **ROLLBACK** (sab undo)
- Sab successful? → **COMMIT** (permanently save)

**Code mein kya kiya?**
1. `Store` banaya → Purane CRUD functions + transaction ki taqat
2. `execTx` banaya → General purpose transaction runner
3. `TransferTx` banaya → Actual money transfer logic (5 steps ek transaction mein)
4. Goroutines se test kiya → 5 parallel transfers fire karke dekha ki code concurrent load jhel raha hai ya nahi

---

## 🗺️ Step-by-Step: Kya Pehle Banao, Kya Baad Mein?

Agar tum ye lecture khud se follow karna chaho toh ye order hai:

### Step 1: SQL Query Files Banao
Pehle humein `entries` aur `transfers` table ke liye SQL queries chahiye (Account ke liye pehle se thi).

📁 **File banao:** `db/query/entry.sql`
```sql
-- name: CreateEntry :one
INSERT INTO entries (account_id, amount)
VALUES ($1, $2)
RETURNING *;

-- name: GetEntry :one
SELECT * FROM entries
WHERE id = $1 LIMIT 1;
```

📁 **File banao:** `db/query/transfer.sql`
```sql
-- name: CreateTransfer :one
INSERT INTO transfers (from_account_id, to_account_id, amount)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers
WHERE id = $1 LIMIT 1;
```

### Step 2: SQLC se Code Generate Karo
```bash
sqlc generate
```
Ye command `db/sqlc/` folder mein `entry.sql.go` aur `transfer.sql.go` auto-generate kar dega.

### Step 3: `main_test.go` Update Karo
`testDB` variable export karo taaki `store_test.go` bhi use kar sake:

📁 **Edit karo:** `db/sqlc/main_test.go`
```go
var testQueries *Queries
var testDB *sql.DB        // <-- ye naya add hua

func TestMain(m *testing.M) {
    var err error
    testDB, err = sql.Open(dbDriver, dbSource)   // conn ki jagah testDB
    // ... baaki same
    testQueries = New(testDB)
}
```

### Step 4: `store.go` Banao (Transaction Logic)
📁 **Naya file banao:** `db/sqlc/store.go`
- `Store` struct define karo (with `*Queries` embedding + `db *sql.DB`)
- `NewStore()` constructor banao
- `execTx()` — generic transaction runner
- `TransferTx()` — actual money transfer logic

### Step 5: `store_test.go` Banao (Concurrent Test)
📁 **Naya file banao:** `db/sqlc/store_test.go`
- `TestTransferTx` function likho
- 5 goroutines fire karo
- Channels se results verify karo

### Step 6: Test Run Karo
```bash
go test -v -cover ./...
```
Sab green aana chahiye! ✅

---

## Ye Lecture Kya Hai?

Ab tak humne sirf single table par CRUD (Create, Read, Update, Delete) operations kiye the. Lekin real-world apps mein (jaise bank app), ek single action (e.g. Money Transfer) ke liye multiple tables update karni padti hain.
Agar unme se ek update fail ho jaye, toh data inconsistent ho jayega. Iss problem ko solve karne ke liye hum **Database Transactions** use karte hain.

Transactions **ACID** properties follow karte hain:
- **A**tomicity: Ya toh saare operations honge, ya ek bhi nahi (Rollback).
- **C**onsistency: Data hamesha valid rules (constraints) follow karega.
- **I**solation: Ek sath chalne wale transactions ek doosre ko effect nahi karenge.
- **D**urability: Transaction successful hone par data permanently save ho jayega.

---

## 🛠 Code Explanation 1: `Store` Struct (`store.go`)

SQLC ne humein `Queries` struct diya tha jo sirf ek single operation run kar sakta tha. Hum ek naya `Store` struct banayenge jiske paas individual queries chalane ki taqat bhi hogi (using Composition) aur naye Transactional functions bhi honge.

```go
type Store struct {
	*Queries          // Composition: Queries ke saare functions isme automatically aa jayenge!
	db *sql.DB        // Transaction shuru karne ke liye raw db connection zaroori hai
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}
```

---

## 🏗 Code Explanation 2: Generic Transaction Executer (`execTx`)

Hum ek common function `execTx` banate hain jo transaction start karega, humara custom kaam (callback function) chalayega, aur uske baad ya toh commit karega ya rollback.

```go
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	// 1. Transaction Start
	tx, err := store.db.BeginTx(ctx, nil)
	
	// 2. Iss naye transaction 'tx' ka ek naya Queries object banao
	q := New(tx)
	
	// 3. Callback function run karo (jisme actual queries run hongi)
	err = fn(q)
	if err != nil {
		// 4. Agar koi error aayi toh pichla sab kuch undo (rollback) kardo
		tx.Rollback()
		return err
	}

	// 5. Agar sab sahi raha toh changes permanently save (commit) kardo
	return tx.Commit()
}
```

---

## 💸 Code Explanation 3: Transfer Transaction (`TransferTx`)

Bank Transfer mein 5 cheezein hoti hain:
1. Transfer record banta hai.
2. FromAccount ki Entry (paisa nikalna) banti hai.
3. ToAccount ki Entry (paisa aana) banti hai.
4. FromAccount ka Balance ghatao. (To be done later)
5. ToAccount ka Balance badhao. (To be done later)

Ye sab hum `execTx` ke callback ke andar chalate hain taaki ye sab ek transaction mein ho:

```go
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		// Step 1: Transfer Record
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{...})
		
		// Step 2: From Entry
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{Amount: -arg.Amount, ...})
		
		// Step 3: To Entry
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{Amount: arg.Amount, ...})

		return nil
	})

	return result, err
}
```

---

## 🧪 Code Explanation 4: Testing with Goroutines (`store_test.go`)

Database transactions sabse zyada fail hote hain concurrency mein (jab bohot saare users ek sath transfer kar rahe ho). Isliye humne test mein `Goroutines` aur `Channels` ka use kiya!

```go
// 5 concurrent (parallel) transfers shuru kiye
n := 5
errs := make(chan error)
results := make(chan TransferTxResult)

for i := 0; i < n; i++ {
	go func() { // Ye background thread me chalega
		result, err := store.TransferTx(context.Background(), ...)
		errs <- err     // Result ko channel ke through waapis bheja
		results <- result
	}()
}

// Phir hum n times channels se data recieve karke verify karte hain
for i := 0; i < n; i++ {
	err := <-errs
	require.NoError(t, err)
	// assertions...
}
```

Agar humara code concurrency jhel gaya iska matlab transaction bilkul safe aur **ACID** compliant hai!
