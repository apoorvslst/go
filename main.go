package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ============================================================
// Database connection string
// ============================================================
const dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
const gormDSN = "host=localhost user=root password=secret dbname=simple_bank port=5432 sslmode=disable"

// ============================================================
// MODELS (shared by all examples)
// ============================================================

// Account represents a bank account
type Account struct {
	ID        int64     `db:"id" json:"id" gorm:"primaryKey;autoIncrement"`
	Owner     string    `db:"owner" json:"owner" gorm:"not null"`
	Balance   int64     `db:"balance" json:"balance" gorm:"not null"`
	Currency  string    `db:"currency" json:"currency" gorm:"not null"`
	CreatedAt time.Time `db:"created_at" json:"created_at" gorm:"not null;default:now()"`
}

func main() {
	fmt.Println("=" + repeat("=", 59))
	fmt.Println("  🏦 Simple Bank — Go Database Examples")
	fmt.Println("=" + repeat("=", 59))
	fmt.Println()

	// Uncomment the example you want to run:
	sqlxExample()
	// gormExample()

	// For SQLC: run `sqlc generate` first, then use the generated code
	// See db/sqlc/ folder after running sqlc generate
}

// ============================================================
// EXAMPLE 1: SQLX (Raw SQL + automatic struct mapping)
// ============================================================
func sqlxExample() {
	fmt.Println("🔵 SQLX Example")
	fmt.Println(repeat("-", 40))

	// Connect
	db, err := sqlx.Connect("postgres", dbSource)
	if err != nil {
		log.Fatal("❌ Cannot connect:", err)
	}
	defer db.Close()
	fmt.Println("✅ Connected to database!")

	ctx := context.Background()

	// CREATE
	var newAccount Account
	err = db.QueryRowxContext(ctx,
		`INSERT INTO accounts (owner, balance, currency) 
		 VALUES ($1, $2, $3) 
		 RETURNING id, owner, balance, currency, created_at`,
		"Apoorv", 1000, "USD",
	).StructScan(&newAccount)
	if err != nil {
		log.Fatal("❌ Create failed:", err)
	}
	fmt.Printf("✅ Created: ID=%d, Owner=%s, Balance=%d %s\n",
		newAccount.ID, newAccount.Owner, newAccount.Balance, newAccount.Currency)

	// READ (one)
	var account Account
	err = db.GetContext(ctx, &account,
		"SELECT * FROM accounts WHERE id = $1", newAccount.ID)
	if err != nil {
		log.Fatal("❌ Get failed:", err)
	}
	fmt.Printf("📖 Got:     ID=%d, Owner=%s, Balance=%d %s\n",
		account.ID, account.Owner, account.Balance, account.Currency)

	// READ (many)
	var accounts []Account
	err = db.SelectContext(ctx, &accounts,
		"SELECT * FROM accounts ORDER BY id LIMIT $1 OFFSET $2", 10, 0)
	if err != nil {
		log.Fatal("❌ List failed:", err)
	}
	fmt.Printf("📋 Found %d account(s)\n", len(accounts))

	// UPDATE
	_, err = db.ExecContext(ctx,
		"UPDATE accounts SET balance = $1 WHERE id = $2",
		2000, newAccount.ID)
	if err != nil {
		log.Fatal("❌ Update failed:", err)
	}
	fmt.Println("✏️  Updated balance to 2000!")

	// Verify the update
	var updated Account
	db.GetContext(ctx, &updated, "SELECT * FROM accounts WHERE id = $1", newAccount.ID)
	fmt.Printf("📖 After update: Balance=%d\n", updated.Balance)

	// DELETE
	_, err = db.ExecContext(ctx,
		"DELETE FROM accounts WHERE id = $1", newAccount.ID)
	if err != nil {
		log.Fatal("❌ Delete failed:", err)
	}
	fmt.Println("🗑️  Deleted account!")
	fmt.Println()
}

// ============================================================
// EXAMPLE 2: GORM (Write Go code, not SQL)
// ============================================================
func gormExample() {
	fmt.Println("🟡 GORM Example")
	fmt.Println(repeat("-", 40))

	// Connect
	db, err := gorm.Open(postgres.Open(gormDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Cannot connect:", err)
	}
	fmt.Println("✅ Connected to database!")

	// CREATE
	newAccount := Account{
		Owner:    "Apoorv",
		Balance:  1000,
		Currency: "USD",
	}
	result := db.Create(&newAccount)
	if result.Error != nil {
		log.Fatal("❌ Create failed:", result.Error)
	}
	fmt.Printf("✅ Created: ID=%d, Owner=%s, Balance=%d %s\n",
		newAccount.ID, newAccount.Owner, newAccount.Balance, newAccount.Currency)

	// READ (one) — find by primary key
	var account Account
	db.First(&account, newAccount.ID)
	fmt.Printf("📖 Got:     ID=%d, Owner=%s, Balance=%d %s\n",
		account.ID, account.Owner, account.Balance, account.Currency)

	// READ (one) — find by condition
	var found Account
	db.Where("owner = ?", "Apoorv").First(&found)
	fmt.Printf("🔍 Found:   ID=%d, Owner=%s\n", found.ID, found.Owner)

	// READ (many)
	var accounts []Account
	db.Limit(10).Offset(0).Order("id").Find(&accounts)
	fmt.Printf("📋 Found %d account(s)\n", len(accounts))

	// UPDATE (one field)
	db.Model(&account).Update("balance", 2000)
	fmt.Println("✏️  Updated balance to 2000!")

	// UPDATE (multiple fields)
	db.Model(&account).Updates(map[string]interface{}{
		"balance":  5000,
		"currency": "INR",
	})
	fmt.Println("✏️  Updated balance=5000, currency=INR!")

	// Verify
	var updated Account
	db.First(&updated, newAccount.ID)
	fmt.Printf("📖 After update: Balance=%d, Currency=%s\n", updated.Balance, updated.Currency)

	// DELETE
	db.Delete(&Account{}, newAccount.ID)
	fmt.Println("🗑️  Deleted account!")
	fmt.Println()
}

// Helper function
func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
