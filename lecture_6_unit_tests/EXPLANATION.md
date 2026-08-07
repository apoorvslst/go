# 🧪 Lecture 6: Unit Testing CRUD Operations in Golang

## Ye Lecture Kya Hai?

Lecture 5 mein humne SQLC se CRUD code generate kiya. Par hume kaise pata chalega ki code sach mein sahi chal raha hai ya nahi?
Iske liye hum **Unit Tests** likhte hain! Is lecture mein hum `testing` package aur `testify` library use karke accounts ki sabhi CRUD operations test karenge.

---

## 🛠 Step 1: Testing ka Setup (`main_test.go`)

Har test ko run karne ke liye hume database connection ki zaroorat padegi. To avoid opening a connection for every single test, hum ek `TestMain` function banate hain, jo test suite start hone se pehle ek baar chalta hai.

`db/sqlc/main_test.go`:
```go
package db

import (
	"database/sql"
	"log"
	"os"
	"testing"
	_ "github.com/lib/pq" // Postgres driver (blank identifier ki wajah se ye hata nahi)
)

var testQueries *Queries // Global object tests ke liye

func TestMain(m *testing.M) {
	conn, err := sql.Open("postgres", "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable")
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(conn)

	// Run all unit tests
	os.Exit(m.Run())
}
```

---

## 🎲 Step 2: Random Data Generator (`util/random.go`)

Tests me hamesha fixed values use karne se future tests crash ho sakte hain (jaise unique constraints tootna). Isliye humne ek `util` package banaya jo random strings, integers aur currency generate karta hai.

```go
package util

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	// Seed ke bina hamesha ek hi pattern aata
	rand.Seed(time.Now().UnixNano()) 
}

func RandomString(n int) string {
	// Generate random characters
}

func RandomMoney() int64 {
	// Generate random money amount
}

func RandomCurrency() string {
	// Returns random from EUR, USD, CAD
}
```

---

## ✅ Step 3: Writing Tests (`account_test.go`)

Golang mein conventions:
- Test files ka naam `_test.go` pe khatam hona chahiye.
- Functions ka naam `Test` se shuru hona chahiye.

Humne assertions ko aasan banane ke liye `github.com/stretchr/testify/require` package use kiya hai.

### Create Account Test ➕
Pehla kaam naya account database mein push karna. Test data generation ke liye hum apne `util` functions use karte hain.

```go
func createRandomAccount(t *testing.T) Account {
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)
	
	// require checks - Agar ye fail hue to aage ka code nahi chalega!
	require.NoError(t, err) 
	require.NotEmpty(t, account)
	require.Equal(t, arg.Owner, account.Owner)
	// etc...

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}
```

### Baki CRUD Tests 🔄
- **GetAccount**: Pehle ek random account create karo, phir uski ID query mein daal kar fetch karo. Phir dono ko `require.Equal()` se compare karo.
- **UpdateAccount**: Account banao, uski `Balance` field ek new random value se update karo. Check karo dono equal aate hain ya nahi.
- **DeleteAccount**: Delete query chalao aur check karo error `nil` hai. Phir us ID ko dobara fetch karne ki koshish karo, iss baar expected result ek error (`sql.ErrNoRows`) hona chahiye!
- **ListAccounts**: Loop laga ke 10 accounts banayo, phir `ListAccounts` mein `limit: 5` aur `offset: 5` set karke 5 items aane ka assert karo (`require.Len(t, accounts, 5)`).

---

## 🏃‍♂️ Step 4: Tests Run Karna (`make test`)

Humne Makefile mein ek nayi command `test` banayi:
```makefile
test:
	go test -v -cover ./...
```
- `-v`: Verbose logs dikhata hai ki kaunsa test paas hua.
- `-cover`: Code coverage dikhata hai (kitne percent code test kiya ja chuka hai).
- `./...`: Project ki saari packages ko run karta hai.

Bas terminal me command chalao:
```bash
make test
```
**Aur sab green hoga! 🎉**
