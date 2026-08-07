package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	// Hum lib/pq use kar rahe hain database se baat karne ke liye
	// Blank identifier (_) ka use isliye hai kyunki hum direct pq ka koi function call nahi karte
	_ "github.com/lib/pq"
)

const (
	// Database ka driver postgres use karenge
	dbDriver = "postgres"
	// Database connection link (URL)
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

// Global variable testQueries - isse hum pure test suite mein DB queries run karenge
var testQueries *Queries

// TestMain main entry point hai is package ke tests ke liye
// m *testing.M test object hai jo tests ko control karta hai
func TestMain(m *testing.M) {
	// 1. Pehle database se connection open karo
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err) // Agar error aaye toh log likho aur tests band kardo
	}

	// 2. testQueries variable ko intialize karo hamare naye connection ke saath
	testQueries = New(conn)

	// 3. m.Run() se saare tests chalao, aur uska jo exit code hoga wahi os.Exit se return kardo (success ya fail)
	os.Exit(m.Run())
}
