# 🛠️ Lecture 4: Generate CRUD Golang Code from SQL

## Ye Lecture Kya Hai?

Is lecture mein sikhaya gaya hai ki **Go mein database se kaise baat karte hain** — matlab CRUD operations:
- **C**reate = Naya data daalo
- **R**ead = Data padho
- **U**pdate = Data badlo
- **D**elete = Data delete karo 

Iske liye **3 popular tarike** hain. Video mein teeno compare kiye gaye hain:

``` 
┌─────────────────────────────────────────────────────────┐
│                                                         │
│   1. SQLC  → Tum SQL likho, ye Go code generate karta   │
│   2. SQLX  → Tum SQL likho, ye struct mein map karta    │
│   3. GORM  → Tum Go likho, ye SQL generate karta        │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 🤔 Pehle Samjho: Standard database/sql Package

Go mein ek built-in package hai `database/sql`. Ye bahut basic hai:

```go
// Standard database/sql — bahut BORING aur TEDIOUS!
rows, err := db.Query("SELECT id, owner, balance FROM accounts WHERE id = $1", 1)

// Ab MANUALLY har column scan karo 😩
var id int64
var owner string
var balance int64
err = rows.Scan(&id, &owner, &balance)
```

**Problems:**
1. Har column ke liye manually variable banana padta hai
2. Agar column ka order change ho jaaye → bug!
3. Agar column ka naam galat likh do → runtime pe pata chalega (build time pe nahi)
4. Bahut saara boilerplate code likhna padta hai

**Isliye log SQLC, SQLX ya GORM use karte hain!**

---

## 🟢 Option 1: SQLC (⭐ RECOMMENDED by the Course!)

### SQLC Kya Karta Hai?

```
Tum SQL query likhte ho → SQLC padh ke Go code generate kar deta hai

Input:  account.sql  (tumhara SQL)
Output: account.sql.go  (auto-generated Go code, type-safe!)
```

### Kyun Best Hai?

1. **Tum sirf SQL likho** — Go code automatically ban jaata hai
2. **Type-safe** — galat type doge toh COMPILE time pe error aayega (runtime pe nahi!)
3. **Fastest** — raw SQL jaata hai, koi middleware nahi
4. **SQL seekhne ki zaroorat hai** — but SQL har jagah kaam aata hai, waste nahi hai

### Step 1: SQLC Install karo

```bash
# Windows
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Mac
brew install sqlc

# Check
sqlc version
```

### Step 2: Config file banao (`sqlc.yaml`)

```yaml
version: "2"
sql:
  - engine: "postgresql"          # Kaunsa database use kar rahe ho
    queries: "db/query/"          # SQL queries YAHAN hain
    schema: "db/migration/"       # Schema (migration files) YAHAN hain
    gen:
      go:
        package: "db"             # Generated Go code ka package name
        out: "db/sqlc"            # Generated code YAHAN save hoga
        emit_json_tags: true      # JSON tags bhi daalo (API ke liye useful)
        emit_prepared_queries: false
        emit_interface: false
        emit_exact_table_names: false
```

### Step 3: SQL Queries likho (`db/query/account.sql`)

```sql
-- name: CreateAccount :one
-- "CreateAccount" = Go function ka naam banega
-- ":one" = ek row return hogi
INSERT INTO accounts (owner, balance, currency)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAccount :one
-- ":one" = ek row return hogi (single account)
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: ListAccounts :many
-- ":many" = bahut saari rows return hongi (slice of accounts)
SELECT * FROM accounts
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateAccount :one
UPDATE accounts
SET balance = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
-- ":exec" = kuch return nahi hoga (bas execute karo)
DELETE FROM accounts
WHERE id = $1;
```

**Magic Comments samjho:**

```
-- name: FunctionName :return_type

:one   → func returns (Account, error)     — EK item
:many  → func returns ([]Account, error)   — BAHUT saare items
:exec  → func returns error                — Bas karo, kuch mat do
```

### Step 4: Code Generate karo

```bash
sqlc generate
```

**Ye kya banta hai?** 3 files automatically!

### Auto-generated file 1: `db/sqlc/models.go`

```go
// YE FILE AUTOMATICALLY BANI HAI — EDIT MAT KARO!
package db

import "time"

// Database ke tables → Go structs ban gaye
type Account struct {
    ID        int64     `json:"id"`
    Owner     string    `json:"owner"`
    Balance   int64     `json:"balance"`
    Currency  string    `json:"currency"`
    CreatedAt time.Time `json:"created_at"`
}

type Entry struct {
    ID        int64     `json:"id"`
    AccountID int64     `json:"account_id"`
    Amount    int64     `json:"amount"`
    CreatedAt time.Time `json:"created_at"`
}

type Transfer struct {
    ID            int64     `json:"id"`
    FromAccountID int64     `json:"from_account_id"`
    ToAccountID   int64     `json:"to_account_id"`
    Amount        int64     `json:"amount"`
    CreatedAt     time.Time `json:"created_at"`
}
```

### Auto-generated file 2: `db/sqlc/db.go`

```go
// Database connection wrapper
package db

import (
    "context"
    "database/sql"
)

type DBTX interface {
    ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
    QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type Queries struct {
    db DBTX
}

func New(db DBTX) *Queries {
    return &Queries{db: db}
}
```

### Auto-generated file 3: `db/sqlc/account.sql.go`

```go
// Tumhare SQL queries → Go functions ban gaye!
package db

import "context"

// ────────── CREATE ──────────
type CreateAccountParams struct {
    Owner    string `json:"owner"`
    Balance  int64  `json:"balance"`
    Currency string `json:"currency"`
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
    row := q.db.QueryRowContext(ctx,
        `INSERT INTO accounts (owner, balance, currency) VALUES ($1, $2, $3) RETURNING *`,
        arg.Owner, arg.Balance, arg.Currency,
    )
    var account Account
    err := row.Scan(&account.ID, &account.Owner, &account.Balance, &account.Currency, &account.CreatedAt)
    return account, err
}

// ────────── READ (one) ──────────
func (q *Queries) GetAccount(ctx context.Context, id int64) (Account, error) {
    row := q.db.QueryRowContext(ctx,
        `SELECT * FROM accounts WHERE id = $1 LIMIT 1`, id,
    )
    var account Account
    err := row.Scan(&account.ID, &account.Owner, &account.Balance, &account.Currency, &account.CreatedAt)
    return account, err
}

// ────────── READ (many) ──────────
type ListAccountsParams struct {
    Limit  int32 `json:"limit"`
    Offset int32 `json:"offset"`
}

func (q *Queries) ListAccounts(ctx context.Context, arg ListAccountsParams) ([]Account, error) {
    rows, err := q.db.QueryContext(ctx,
        `SELECT * FROM accounts ORDER BY id LIMIT $1 OFFSET $2`,
        arg.Limit, arg.Offset,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var accounts []Account
    for rows.Next() {
        var account Account
        if err := rows.Scan(&account.ID, &account.Owner, &account.Balance, &account.Currency, &account.CreatedAt); err != nil {
            return nil, err
        }
        accounts = append(accounts, account)
    }
    return accounts, nil
}

// ────────── DELETE ──────────
func (q *Queries) DeleteAccount(ctx context.Context, id int64) error {
    _, err := q.db.ExecContext(ctx, `DELETE FROM accounts WHERE id = $1`, id)
    return err
}
```

### Step 5: Use karo!

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"

    _ "github.com/lib/pq"
    db "github.com/apoor/simple_bank/db/sqlc"
)

func main() {
    // Database se connect karo
    conn, err := sql.Open("postgres",
        "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }

    // SQLC ka queries object banao
    queries := db.New(conn)
    ctx := context.Background()

    // CREATE — Naya account banao
    account, _ := queries.CreateAccount(ctx, db.CreateAccountParams{
        Owner:    "Apoorv",
        Balance:  1000,
        Currency: "INR",
    })
    fmt.Println("Bana diya:", account)

    // READ — Account padho
    got, _ := queries.GetAccount(ctx, account.ID)
    fmt.Println("Mil gaya:", got)

    // LIST — Saare accounts
    all, _ := queries.ListAccounts(ctx, db.ListAccountsParams{Limit: 10, Offset: 0})
    fmt.Println("Saare:", all)

    // DELETE — Account delete karo
    queries.DeleteAccount(ctx, account.ID)
    fmt.Println("Delete ho gaya!")
}
```

---

## 🔵 Option 2: SQLX (Raw SQL + Struct Mapping)

### SQLX Kya Karta Hai?

```
SQLX = database/sql + SUPERPOWERS

database/sql: Manually har column scan karo 😩
SQLX:         Automatically struct mein daal do 😎
```

### SQLX mein kuch generate nahi hota — tum sab khud likhte ho

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

// Struct KHUD banana padta hai (SQLC mein auto banta hai)
type Account struct {
    ID        int64     `db:"id"`          // `db` tag = column name
    Owner     string    `db:"owner"`
    Balance   int64     `db:"balance"`
    Currency  string    `db:"currency"`
    CreatedAt time.Time `db:"created_at"`
}

func main() {
    // ──── CONNECT ────
    // sqlx.Connect use karo (sql.Open nahi)
    db, err := sqlx.Connect("postgres",
        "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    ctx := context.Background()

    // ──── CREATE ────
    var newAccount Account
    err = db.QueryRowxContext(ctx,
        `INSERT INTO accounts (owner, balance, currency) 
         VALUES ($1, $2, $3) 
         RETURNING *`,
        "Apoorv", 1000, "INR",
    ).StructScan(&newAccount)    // ← YE HAI SUPERPOWER! Auto mapping!
    fmt.Println("Bana diya:", newAccount)

    // ──── READ (one) ────
    var account Account
    err = db.GetContext(ctx, &account,
        "SELECT * FROM accounts WHERE id = $1",
        newAccount.ID,
    )
    // db.Get = 1 row → 1 struct (automatically!)
    fmt.Println("Mil gaya:", account)

    // ──── READ (many) ────
    var accounts []Account
    err = db.SelectContext(ctx, &accounts,
        "SELECT * FROM accounts ORDER BY id LIMIT $1 OFFSET $2",
        10, 0,
    )
    // db.Select = many rows → slice of structs (automatically!)
    fmt.Println("Saare:", len(accounts), "accounts")

    // ──── UPDATE ────
    _, err = db.ExecContext(ctx,
        "UPDATE accounts SET balance = $1 WHERE id = $2",
        5000, newAccount.ID,
    )
    fmt.Println("Balance update ho gaya!")

    // ──── DELETE ────
    _, err = db.ExecContext(ctx,
        "DELETE FROM accounts WHERE id = $1",
        newAccount.ID,
    )
    fmt.Println("Delete ho gaya!")

    // ──── NAMED QUERIES (SQLX ka special feature!) ────
    // $1, $2 ke badle :fieldname use karo
    _, err = db.NamedExecContext(ctx,
        `INSERT INTO accounts (owner, balance, currency) 
         VALUES (:owner, :balance, :currency)`,
        Account{Owner: "Bob", Balance: 500, Currency: "USD"},
    )
    // :owner → Account.Owner se match hoga
    // :balance → Account.Balance se match hoga
    // Bahut readable hai!
    fmt.Println("Bob ka account bhi ban gaya!")
}
```

### SQLX Important Functions

```
┌─────────────────────┬────────────────────────────────────────┐
│ Function            │ Kya karta hai                          │
├─────────────────────┼────────────────────────────────────────┤
│ sqlx.Connect()      │ Database se connect karo               │
│ db.GetContext()     │ 1 row → 1 struct mein daalo            │
│ db.SelectContext()  │ Many rows → slice mein daalo           │
│ db.ExecContext()    │ Query chalao, kuch return mat karo     │
│ db.QueryRowx()     │ 1 row query (manual scan bhi kar sakte)│
│ db.Queryx()        │ Many rows (manual scan)                │
│ db.NamedExec()     │ :fieldname wali query                  │
│ .StructScan()      │ Row ko struct mein convert karo        │
└─────────────────────┴────────────────────────────────────────┘
```

---

## 🟡 Option 3: GORM (Full ORM — Go Code Likho, SQL Mat Likho)

### GORM Kya Karta Hai?

```
GORM mein tum Go code likhte ho, GORM khud SQL generate karta hai

Tum likhte ho:  db.Create(&account)
GORM chalata hai: INSERT INTO accounts (owner, balance, currency) VALUES (...)

Tum likhte ho:  db.First(&account, 1)
GORM chalata hai: SELECT * FROM accounts WHERE id = 1 LIMIT 1

Tum likhte ho:  db.Delete(&Account{}, 1)
GORM chalata hai: DELETE FROM accounts WHERE id = 1
```

### GORM ka Code

```go
package main

import (
    "fmt"
    "log"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// Model define karo (struct tags se GORM samajhta hai table structure)
type Account struct {
    ID        int64     `gorm:"primaryKey;autoIncrement"`
    Owner     string    `gorm:"not null"`
    Balance   int64     `gorm:"not null"`
    Currency  string    `gorm:"not null"`
    CreatedAt time.Time `gorm:"not null;default:now()"`
}

func main() {
    // ──── CONNECT ────
    dsn := "host=localhost user=root password=secret dbname=simple_bank port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // ──── AUTO MIGRATE (optional) ────
    // Ye struct dekh ke table bana deta hai / update kar deta hai
    // Production mein mat karo! golang-migrate use karo.
    // db.AutoMigrate(&Account{})

    // ──── CREATE ────
    newAccount := Account{
        Owner:    "Apoorv",
        Balance:  1000,
        Currency: "INR",
    }
    db.Create(&newAccount)
    // Create ke baad newAccount.ID automatically fill ho jaata hai!
    fmt.Println("Bana diya! ID:", newAccount.ID)

    // ──── READ (one) — by primary key ────
    var account Account
    db.First(&account, newAccount.ID)  // Find by ID
    fmt.Println("Mil gaya:", account)

    // ──── READ (one) — by condition ────
    var found Account
    db.Where("owner = ?", "Apoorv").First(&found)
    fmt.Println("Naam se mila:", found)

    // ──── READ (many) ────
    var accounts []Account
    db.Limit(10).Offset(0).Order("id").Find(&accounts)
    fmt.Println("Saare:", len(accounts), "accounts")

    // ──── READ with multiple conditions ────
    var usdAccounts []Account
    db.Where("currency = ? AND balance > ?", "USD", 100).Find(&usdAccounts)

    // ──── UPDATE (ek field) ────
    db.Model(&account).Update("balance", 5000)
    fmt.Println("Balance update hua!")

    // ──── UPDATE (multiple fields) ────
    db.Model(&account).Updates(map[string]interface{}{
        "balance":  10000,
        "currency": "USD",
    })
    fmt.Println("Multiple fields update hue!")

    // ──── DELETE ────
    db.Delete(&Account{}, newAccount.ID)
    fmt.Println("Delete ho gaya!")

    // ──── TRANSACTION (Bank transfer ke liye ZARURI!) ────
    // Apoorv → Bob ko 500 rupaye
    err = db.Transaction(func(tx *gorm.DB) error {
        // Step 1: Apoorv se 500 ghataao
        if err := tx.Model(&Account{}).Where("id = ?", 1).
            Update("balance", gorm.Expr("balance - ?", 500)).Error; err != nil {
            return err  // ERROR? Toh sab undo (rollback)!
        }

        // Step 2: Bob mein 500 badhaao
        if err := tx.Model(&Account{}).Where("id = ?", 2).
            Update("balance", gorm.Expr("balance + ?", 500)).Error; err != nil {
            return err  // ERROR? Toh sab undo (rollback)!
        }

        // Step 3: Transfer record banao
        if err := tx.Create(&Transfer{
            FromAccountID: 1,
            ToAccountID:   2,
            Amount:        500,
        }).Error; err != nil {
            return err  // ERROR? Toh sab undo (rollback)!
        }

        return nil  // Sab theek? COMMIT! (permanently save)
    })

    if err != nil {
        fmt.Println("Transfer FAIL hua, sab undo ho gaya!")
    } else {
        fmt.Println("Transfer SUCCESS!")
    }
}

// Transfer model (GORM ke liye)
type Transfer struct {
    ID            int64     `gorm:"primaryKey;autoIncrement"`
    FromAccountID int64     `gorm:"not null"`
    ToAccountID   int64     `gorm:"not null"`
    Amount        int64     `gorm:"not null"`
    CreatedAt     time.Time `gorm:"not null;default:now()"`
}
```

### GORM Methods Cheat Sheet

```
┌───────────────────────────────┬──────────────────────────────────────┐
│ GORM Code                     │ SQL mein kya chalega                 │
├───────────────────────────────┼──────────────────────────────────────┤
│ db.Create(&account)           │ INSERT INTO accounts ...             │
│ db.First(&account, id)        │ SELECT * WHERE id=? LIMIT 1         │
│ db.Find(&accounts)            │ SELECT * FROM accounts               │
│ db.Where("x=?", v).First()   │ SELECT * WHERE x=v LIMIT 1          │
│ db.Where("x=?", v).Find()    │ SELECT * WHERE x=v                   │
│ db.Model(&a).Update("k", v)  │ UPDATE accounts SET k=v WHERE...     │
│ db.Model(&a).Updates(map)     │ UPDATE accounts SET k=v, k2=v2...   │
│ db.Delete(&Account{}, id)     │ DELETE FROM accounts WHERE id=?     │
│ db.Limit(10).Offset(5)       │ ... LIMIT 10 OFFSET 5                │
│ db.Order("id desc")          │ ... ORDER BY id DESC                 │
│ db.Transaction(func...)       │ BEGIN; ...; COMMIT or ROLLBACK       │
└───────────────────────────────┴──────────────────────────────────────┘
```

---

## 🏆 Comparison: Kaunsa Use Karu?

### Quick Comparison Table

| Feature | SQLC | SQLX | GORM |
|---------|------|------|------|
| **Tum kya likhte ho** | SQL | SQL | Go code |
| **Code generate hota hai?** | ✅ Haan | ❌ Nahi | ❌ Nahi |
| **Type safety** | ✅ Compile time | ⚠️ Runtime | ⚠️ Runtime |
| **Speed** | 🚀 Sabse fast | ⚡ Fast | 🐢 Sabse slow |
| **SQL control** | Full | Full | Limited |
| **Seekhne mein** | Easy | Medium | Easy start, hard later |
| **Complex queries** | Easy (SQL hi hai) | Easy (SQL hi hai) | Hard (GORM syntax) |
| **Error kab milega** | Build time ✅ | Runtime ❌ | Runtime ❌ |

### Analogy (Real Life Example)

```
SQLC   = Restaurant mein tum RECIPE likho, chef (SQLC) dish bana de
         → Tum SQL likho, SQLC Go code bana de
         → Tumhe cooking nahi aani chahiye (Go boilerplate)

SQLX   = Restaurant mein tum khud cook karo, but kitchen mein 
         smart tools hain jo kaam easy kar de
         → Tum SQL + Go dono likho, but helpers milte hain

GORM   = Swiggy/Zomato — tum bas button dabao, khaana aa jaaye
         → Tum Go likho, SQL ki tension nahi
         → But customization limited hai
```

### Kab Kya Use Karo?

```
✅ SQLC use karo jab:
   - Production app bana rahe ho
   - Performance chahiye
   - SQL aata hai ya seekhna chahte ho
   - Compile-time safety chahiye (bugs early pakadna)

✅ SQLX use karo jab:
   - Dynamic queries chahiye (runtime pe query change ho)
   - Flexibility chahiye
   - SQL bahut accha aata hai

✅ GORM use karo jab:
   - SQL nahi aata / seekhna nahi chahte
   - Quick prototype bana rahe ho
   - Simple CRUD hi karna hai
   ⚠️ Complex queries (joins, subqueries) mein mushkil hoga
```

---

## 📁 Is Folder Mein Kya Hai

```
lecture_4_crud_golang_sqlc_sqlx_gorm/
├── EXPLANATION.md              ← Ye file (detailed explanation)
├── sqlc_example/
│   ├── sqlc.yaml               ← SQLC config
│   └── db/
│       └── query/
│           └── account.sql     ← SQL queries (SQLC input)
├── sqlx_example/
│   └── main.go                 ← SQLX ka complete working code
└── gorm_example/
    └── main.go                 ← GORM ka complete working code
```

---

## 💡 Key Takeaways

1. **CRUD = Create, Read, Update, Delete** — ye 4 operations sab kuch cover karte hain
2. **SQLC = BEST choice** (course creator ka bhi yahi recommendation hai)
3. **SQLC mein tum SQL likho → Go code auto-generate hota hai** (no manual mapping!)
4. **SQLX mein tum SQL + Go dono likho** but `db.Get()` aur `db.Select()` life easy karte hain
5. **GORM mein SQL mat likho** — Go code se kaam ho jaata hai (but slow hai aur complex queries mein mushkil)
6. **Transaction** bahut important hai bank app mein — "ya toh sab ho, ya kuch na ho"
7. **`:one`, `:many`, `:exec`** — SQLC ke magic comments yaad rakhna!
