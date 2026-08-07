# 🛠️ Lecture 5: CRUD Operations with SQLC

## Ye Lecture Kya Hai?

Is lecture mein hum seekhte hain ki kaise hum apni Simple Bank application ke liye **Golang mein CRUD operations** (Create, Read, Update, Delete) implement kar sakte hain.

Humne 4 options explore kiye:
1. `database/sql` (Standard library) -> Fast, but code lamba aur error-prone hota hai.
2. `GORM` -> ORM hai, likhna aasan hai but slow aur complex queries me mushkil hoti hai.
3. `sqlx` -> Achha hai, map apne aap karta hai par queries manual reh jaati hain.
4. **`sqlc` (The Winner!) 🏆** -> Fastest aur sabse safe!

SQLC mein hum kya karte hain?
```
Tum sirf SQL queries likho → SQLC automatically tumhare liye Golang ka type-safe CRUD code generate kar dega!
```

---

## 🚀 Step 1: SQLC ka Setup (Config)

SQLC ko batana padta hai ki humara database kaisa hai aur code kahan generate karna hai. Ye hum `sqlc.yaml` file se karte hain:

```yaml
version: "2"
sql:
  - engine: "postgresql"          # Database engine
    queries: "db/query/"          # Yahan hum apne raw SQL queries rakhenge
    schema: "db/migration/"       # Yahan tables ke schema (migration files) hain
    gen:
      go:
        package: "db"             # Generated code ka Go package
        out: "db/sqlc"            # Generated files is folder mein save hongi
        emit_json_tags: true      # Structs mein JSON tags lagao (APIs ke liye zaroori)
        emit_prepared_queries: false
        emit_interface: false
        emit_exact_table_names: false
```

---

## 📝 Step 2: Queries Likhna (`account.sql`)

SQLC ka magic comments mein chupa hai! Ek simple `.sql` file banti hai aur SQLC usko Go functions bana deta hai:

```sql
-- name: CreateAccount :one
-- ":one" ka matlab hai ki 1 row wapis aayegi (Account object)
INSERT INTO accounts (owner, balance, currency)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: ListAccounts :many
-- ":many" ka matlab slice of records wapis aayega ([]Account)
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
-- ":exec" ka matlab function sirf execute hoga (returns only error)
DELETE FROM accounts
WHERE id = $1;
```

---

## 🪄 Step 3: Magic in Action (`make sqlc`)

Jese hi tum terminal mein run karoge:
```bash
make sqlc
```
Ya direct `sqlc generate` chalane par, SQLC 3 Go files banata hai `db/sqlc/` folder mein. Tumhe in files ko kabhi khud se edit nahi karna hai!

### 1️⃣ `models.go` (Tables ke Structs)
Ye DB schema ko padh kar automatically structs bana deta hai:
```go
package db

import (
	"time"
)

type Account struct {
	ID        int64     `json:"id"`
	Owner     string    `json:"owner"`
	Balance   int64     `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}
// Entry aur Transfer ke liye bhi aise hi banega...
```

### 2️⃣ `db.go` (Connection Setup)
Ye file connection ko sambhalti hai jiske andar ek `Queries` struct banta hai:
```go
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}
```

### 3️⃣ `account.sql.go` (The Actual CRUD Functions)
Ye sabse important file hai. Yahan SQLC ne tumhari `.sql` file ko Go Code mein badal diya:

#### **Create (Generated Code)**:
```go
type CreateAccountParams struct {
	Owner    string `json:"owner"`
	Balance  int64  `json:"balance"`
	Currency string `json:"currency"`
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, createAccount, arg.Owner, arg.Balance, arg.Currency)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Owner,
		&i.Balance,
		&i.Currency,
		&i.CreatedAt,
	)
	return i, err
}
```
*Notice karo ki yahan automatically `RETURNING *` ko explicitly `RETURNING id, owner, balance...` mein convert kiya gaya hai, jo ki bugs avoid karne mein madad karta hai.*

#### **Read Multiple (Generated Code)**:
```go
type ListAccountsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListAccounts(ctx context.Context, arg ListAccountsParams) ([]Account, error) {
	rows, err := q.db.QueryContext(ctx, listAccounts, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(&i.ID, &i.Owner, &i.Balance, &i.Currency, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	// ... errors handle karke finally items return
	return items, nil
}
```

---

## 🎯 Conclusion: SQLC Kyun?

1. **Typos se azadi**: Agar SQL mein koi typo hoga (jaise `accounts` ki jagah `account`), toh SQLC code generate hi nahi karega! Error terminal par dikhega, runtime par app crash nahi hogi.
2. **Speed**: Go ka fast `database/sql` use karta hai under the hood, koi unnecessary wrapper nahi.
3. **Maza**: Code manually likhna nahi padta, DB schema update karo, SQL likho, aur baaki kaam SQLC ko de do!
