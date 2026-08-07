package main

// ============================================================
// SQLX Example: Raw SQL + Automatic Struct Mapping
//
// SQLX = database/sql ka upgraded version
// Fayda: db.Get() aur db.Select() se struct mein auto map hota hai
// ============================================================

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver (blank import)
)

// Account struct — KHUD banana padta hai (SQLC mein auto banta)
// `db:"column_name"` tag se SQLX ko batate hain kaunsa column kaunse field mein jaaye
type Account struct {
	ID        int64     `db:"id"`
	Owner     string    `db:"owner"`
	Balance   int64     `db:"balance"`
	Currency  string    `db:"currency"`
	CreatedAt time.Time `db:"created_at"`
}

func main() {
	fmt.Println("🔵 SQLX Example")
	fmt.Println("────────────────────────────────────────")

	// ──────────────────────────────────────────────────
	// STEP 1: Database se CONNECT karo
	// ──────────────────────────────────────────────────
	// sqlx.Connect use karo (NOT sql.Open)
	// ye internally sql.Open + db.Ping dono karta hai
	db, err := sqlx.Connect("postgres",
		"postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable")
	if err != nil {
		log.Fatal("❌ Connect nahi ho paaya:", err)
	}
	defer db.Close() // Function end pe connection band karo
	fmt.Println("✅ Database se connect ho gaye!")

	ctx := context.Background()

	// ──────────────────────────────────────────────────
	// STEP 2: CREATE — Naya account banao
	// ──────────────────────────────────────────────────
	var newAccount Account
	err = db.QueryRowxContext(ctx,
		// Raw SQL query — tum control mein ho!
		`INSERT INTO accounts (owner, balance, currency) 
		 VALUES ($1, $2, $3) 
		 RETURNING id, owner, balance, currency, created_at`,
		"Apoorv", 1000, "INR", // $1=Apoorv, $2=1000, $3=INR
	).StructScan(&newAccount) // ← SUPERPOWER: auto mapping!
	// StructScan automatically columns ko struct fields mein daal deta hai
	// id → Account.ID, owner → Account.Owner, etc.

	if err != nil {
		log.Fatal("❌ Account nahi bana:", err)
	}
	fmt.Printf("✅ CREATE: ID=%d, Owner=%s, Balance=%d %s\n",
		newAccount.ID, newAccount.Owner, newAccount.Balance, newAccount.Currency)

	// ──────────────────────────────────────────────────
	// STEP 3: READ (one) — Ek account padho
	// ──────────────────────────────────────────────────
	var account Account
	err = db.GetContext(ctx, &account,
		"SELECT * FROM accounts WHERE id = $1",
		newAccount.ID,
	)
	// db.Get = 1 row fetch karo aur struct mein daalo
	// Internally ye: QueryRow → Scan → struct mein daal diya

	if err != nil {
		log.Fatal("❌ Account nahi mila:", err)
	}
	fmt.Printf("📖 READ:   ID=%d, Owner=%s, Balance=%d %s\n",
		account.ID, account.Owner, account.Balance, account.Currency)

	// ──────────────────────────────────────────────────
	// STEP 4: READ (many) — Saare accounts padho
	// ──────────────────────────────────────────────────
	var accounts []Account
	err = db.SelectContext(ctx, &accounts,
		"SELECT * FROM accounts ORDER BY id LIMIT $1 OFFSET $2",
		10, 0, // 10 results, page 1 (offset 0)
	)
	// db.Select = multiple rows → slice of structs
	// Ye sabse useful function hai SQLX mein!

	if err != nil {
		log.Fatal("❌ Accounts nahi mile:", err)
	}
	fmt.Printf("📋 LIST:   %d account(s) mile\n", len(accounts))
	for _, a := range accounts {
		fmt.Printf("   → ID=%d, Owner=%s, Balance=%d %s\n",
			a.ID, a.Owner, a.Balance, a.Currency)
	}

	// ──────────────────────────────────────────────────
	// STEP 5: UPDATE — Balance change karo
	// ──────────────────────────────────────────────────
	_, err = db.ExecContext(ctx,
		"UPDATE accounts SET balance = $1 WHERE id = $2",
		5000, newAccount.ID,
	)
	if err != nil {
		log.Fatal("❌ Update nahi hua:", err)
	}

	// Verify: updated value padho
	var updated Account
	db.GetContext(ctx, &updated, "SELECT * FROM accounts WHERE id = $1", newAccount.ID)
	fmt.Printf("✏️  UPDATE: Balance %d → %d\n", newAccount.Balance, updated.Balance)

	// ──────────────────────────────────────────────────
	// STEP 6: NAMED QUERY — SQLX ka special feature!
	// ──────────────────────────────────────────────────
	// $1, $2, $3 ke badle :owner, :balance, :currency use karo
	// Struct ke field names se match hota hai — bahut readable!
	_, err = db.NamedExecContext(ctx,
		`INSERT INTO accounts (owner, balance, currency) 
		 VALUES (:owner, :balance, :currency)`,
		Account{Owner: "Bob", Balance: 500, Currency: "USD"},
		// :owner  → Account.Owner  → "Bob"
		// :balance → Account.Balance → 500
		// :currency → Account.Currency → "USD"
	)
	if err != nil {
		log.Fatal("❌ Named query fail:", err)
	}
	fmt.Println("✅ Named query se Bob ka account ban gaya!")

	// ──────────────────────────────────────────────────
	// STEP 7: DELETE — Account delete karo
	// ──────────────────────────────────────────────────
	_, err = db.ExecContext(ctx,
		"DELETE FROM accounts WHERE id = $1",
		newAccount.ID,
	)
	if err != nil {
		log.Fatal("❌ Delete nahi hua:", err)
	}
	fmt.Println("🗑️  DELETE: Account delete ho gaya!")

	// Bob ka bhi delete karo (cleanup)
	db.ExecContext(ctx, "DELETE FROM accounts WHERE owner = $1", "Bob")

	fmt.Println("\n✅ SQLX example complete!")
}
