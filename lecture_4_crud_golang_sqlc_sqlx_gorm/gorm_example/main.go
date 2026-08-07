package main

// ============================================================
// GORM Example: Full ORM — Go Code Likho, SQL Mat Likho
//
// GORM = Go Object-Relational Mapping
// Fayda: SQL bilkul nahi likhna padta, Go methods se sab hota hai
// Nuksan: Complex queries mein mushkil, aur thoda slow
// ============================================================

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Account model — GORM isse read karke table structure samajhta hai
// Struct tags se constraints define hote hain
type Account struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"` // Primary key, auto increment
	Owner     string    `gorm:"not null"`                 // NULL nahi ho sakta
	Balance   int64     `gorm:"not null"`
	Currency  string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`   // Default = current time
}

// Entry model — foreign key ke saath!
type Entry struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	AccountID int64     `gorm:"not null"`          // Foreign key column
	Account   Account   `gorm:"foreignKey:AccountID"` // Relationship define karo
	Amount    int64     `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

// Transfer model
type Transfer struct {
	ID            int64     `gorm:"primaryKey;autoIncrement"`
	FromAccountID int64     `gorm:"not null"`
	FromAccount   Account   `gorm:"foreignKey:FromAccountID"` // Sender
	ToAccountID   int64     `gorm:"not null"`
	ToAccount     Account   `gorm:"foreignKey:ToAccountID"`   // Receiver
	Amount        int64     `gorm:"not null"`
	CreatedAt     time.Time `gorm:"not null;default:now()"`
}

func main() {
	fmt.Println("🟡 GORM Example")
	fmt.Println("────────────────────────────────────────")

	// ──────────────────────────────────────────────────
	// STEP 1: Database se CONNECT karo
	// ──────────────────────────────────────────────────
	// GORM ka apna connection format hai (DSN = Data Source Name)
	dsn := "host=localhost user=root password=secret dbname=simple_bank port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Connect nahi ho paaya:", err)
	}
	fmt.Println("✅ Database se connect ho gaye!")

	// ──────────────────────────────────────────────────
	// AUTO MIGRATE (⚠️ Sirf development mein use karo!)
	// ──────────────────────────────────────────────────
	// Ye struct dekh ke table bana deta hai ya update kar deta hai
	// Production mein golang-migrate use karo, ye nahi!
	// db.AutoMigrate(&Account{}, &Entry{}, &Transfer{})

	// ──────────────────────────────────────────────────
	// STEP 2: CREATE — Naya account banao
	// ──────────────────────────────────────────────────
	newAccount := Account{
		Owner:    "Apoorv",
		Balance:  1000,
		Currency: "INR",
	}
	result := db.Create(&newAccount)
	// IMPORTANT: Create ke baad newAccount.ID automatically fill ho jaata hai!
	// Ye GORM ka magic hai — INSERT ke baad RETURNING * chala ke ID set karta hai

	if result.Error != nil {
		log.Fatal("❌ Account nahi bana:", result.Error)
	}
	fmt.Printf("✅ CREATE: ID=%d, Owner=%s, Balance=%d %s\n",
		newAccount.ID, newAccount.Owner, newAccount.Balance, newAccount.Currency)
	fmt.Printf("   (Rows affected: %d)\n", result.RowsAffected)

	// ──────────────────────────────────────────────────
	// STEP 3: READ (one) — Primary Key se dhoondho
	// ──────────────────────────────────────────────────
	var account Account
	db.First(&account, newAccount.ID) // First = LIMIT 1
	// Internally GORM ye SQL chalata hai:
	// SELECT * FROM accounts WHERE id = 1 ORDER BY id LIMIT 1
	fmt.Printf("📖 READ (by ID):   Owner=%s, Balance=%d\n", account.Owner, account.Balance)

	// ──────────────────────────────────────────────────
	// STEP 4: READ (one) — Condition se dhoondho
	// ──────────────────────────────────────────────────
	var found Account
	db.Where("owner = ?", "Apoorv").First(&found)
	// SQL: SELECT * FROM accounts WHERE owner = 'Apoorv' LIMIT 1
	fmt.Printf("📖 READ (by name): Owner=%s, Balance=%d\n", found.Owner, found.Balance)

	// ──────────────────────────────────────────────────
	// STEP 5: READ (many) — Multiple accounts
	// ──────────────────────────────────────────────────
	var accounts []Account
	db.Limit(10).Offset(0).Order("id").Find(&accounts)
	// SQL: SELECT * FROM accounts ORDER BY id LIMIT 10 OFFSET 0
	fmt.Printf("📋 LIST: %d account(s) mile\n", len(accounts))
	for _, a := range accounts {
		fmt.Printf("   → ID=%d, Owner=%s, Balance=%d %s\n",
			a.ID, a.Owner, a.Balance, a.Currency)
	}

	// ──────────────────────────────────────────────────
	// STEP 5b: READ with multiple conditions
	// ──────────────────────────────────────────────────
	var richAccounts []Account
	db.Where("balance > ? AND currency = ?", 500, "INR").Find(&richAccounts)
	// SQL: SELECT * FROM accounts WHERE balance > 500 AND currency = 'INR'
	fmt.Printf("💰 Rich INR accounts: %d\n", len(richAccounts))

	// ──────────────────────────────────────────────────
	// STEP 6: UPDATE — Ek field change karo
	// ──────────────────────────────────────────────────
	db.Model(&account).Update("balance", 5000)
	// SQL: UPDATE accounts SET balance = 5000 WHERE id = ?
	fmt.Println("✏️  UPDATE (one field): Balance → 5000")

	// ──────────────────────────────────────────────────
	// STEP 7: UPDATE — Multiple fields change karo
	// ──────────────────────────────────────────────────

	// Method 1: Map use karo
	db.Model(&account).Updates(map[string]interface{}{
		"balance":  10000,
		"currency": "USD",
	})
	fmt.Println("✏️  UPDATE (map): Balance=10000, Currency=USD")

	// Method 2: Struct use karo (⚠️ WARNING: zero values skip ho jaate hain!)
	// db.Model(&account).Updates(Account{Balance: 0})
	// ↑ YE KAAM NAHI KAREGA! Balance=0 ko GORM skip kar dega
	//   kyunki Go mein 0 = zero value = "empty" samjhta hai
	//   Isliye map use karo agar zero values set karne hain

	// Verify the update
	var updated Account
	db.First(&updated, newAccount.ID)
	fmt.Printf("📖 After update: Balance=%d, Currency=%s\n", updated.Balance, updated.Currency)

	// ──────────────────────────────────────────────────
	// STEP 8: TRANSACTION — Sabse important bank ke liye!
	// ──────────────────────────────────────────────────
	//
	// Transaction kya hai?
	// "Ya toh SAB karo, ya KUCH MAT karo"
	//
	// Example: Apoorv → Bob ko 500 rupaye
	// Step 1: Apoorv se 500 ghataao
	// Step 2: Bob mein 500 badhaao
	//
	// Agar Step 1 ho jaaye but Step 2 fail ho jaaye?
	// DISASTER! Paisa gayab ho jaayega!
	// Transaction ensure karta hai ki ya toh dono steps ho, ya koi nahi

	// Pehle Bob ka account banate hain
	bob := Account{Owner: "Bob", Balance: 2000, Currency: "USD"}
	db.Create(&bob)

	fmt.Println("\n💸 Transaction: Apoorv → Bob, 500 rupaye...")
	err = db.Transaction(func(tx *gorm.DB) error {
		// tx = transaction object (db ke badle tx use karo!)

		// Step 1: Apoorv se 500 ghataao
		if err := tx.Model(&Account{}).
			Where("id = ?", newAccount.ID).
			Update("balance", gorm.Expr("balance - ?", 500)).Error; err != nil {
			return err // ← ERROR → ROLLBACK (sab undo!)
		}
		fmt.Println("   ✓ Apoorv se 500 ghate")

		// Step 2: Bob mein 500 badhaao
		if err := tx.Model(&Account{}).
			Where("id = ?", bob.ID).
			Update("balance", gorm.Expr("balance + ?", 500)).Error; err != nil {
			return err // ← ERROR → ROLLBACK (Step 1 bhi undo!)
		}
		fmt.Println("   ✓ Bob mein 500 badhe")

		// Step 3: Transfer record banao
		if err := tx.Create(&Transfer{
			FromAccountID: newAccount.ID,
			ToAccountID:   bob.ID,
			Amount:        500,
		}).Error; err != nil {
			return err // ← ERROR → ROLLBACK (Step 1 & 2 bhi undo!)
		}
		fmt.Println("   ✓ Transfer record bana")

		return nil // ← SAB THEEK → COMMIT (permanently save!)
	})

	if err != nil {
		fmt.Println("❌ Transfer FAIL! Sab undo ho gaya (rollback)")
	} else {
		fmt.Println("✅ Transfer SUCCESS! (committed)")
	}

	// Check final balances
	var apoorvFinal, bobFinal Account
	db.First(&apoorvFinal, newAccount.ID)
	db.First(&bobFinal, bob.ID)
	fmt.Printf("   Apoorv: %d, Bob: %d\n", apoorvFinal.Balance, bobFinal.Balance)

	// ──────────────────────────────────────────────────
	// STEP 9: DELETE
	// ──────────────────────────────────────────────────

	// Pehle transfers delete karo (foreign key constraint!)
	db.Where("from_account_id = ? OR to_account_id = ?", newAccount.ID, newAccount.ID).
		Delete(&Transfer{})

	// Phir accounts delete karo
	db.Delete(&Account{}, newAccount.ID)
	db.Delete(&Account{}, bob.ID)
	fmt.Println("\n🗑️  DELETE: Saare test accounts delete ho gaye!")

	fmt.Println("\n✅ GORM example complete!")
}
