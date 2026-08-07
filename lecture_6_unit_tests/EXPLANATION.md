# 🧪 Lecture 6: Unit Testing CRUD Operations in Golang

---

## 🦅 High-Level Overview (Bird's Eye View)

Abhi tak humne Database banaya (Lecture 2), Migration script likhi (Lecture 3), aur Golang mein CRUD code generate kiya (Lecture 5). Lekin humein kaise pata chalega ki jo code humne likha hai wo sach mein Database se sahi baat kar raha hai?

**Solution: Unit Tests**
Unit Tests ek automated tareeka hai apne code ko check karne ka. Isme hum ek program likhte hain jo humare actual program ko test karta hai! Agar koi change code break karta hai, toh test fail ho jayega aur humein deploy karne se pehle hi pata chal jayega.

Golang mein testing inbuilt aati hai (`testing` package), lekin assertions (check karna ki value sahi aayi ya nahi) thoda lamba hota hai. Isliye hum ek library use karenge: **Testify**.

**Testing ka flow kya hoga?**
1. Test start hone se pehle database connection open hoga (`main_test.go`).
2. Har test ke liye hum random data generate karenge taaki puraane tests naye tests ke saath conflict na karein (`util/random.go`).
3. Hum Create, Read, Update aur Delete ke functions chala kar dekhenge ki database sahi results wapis bhej raha hai ya nahi (`account_test.go`).

---

## 🗺️ Step-by-Step: Kya Pehle Banao, Kya Baad Mein?

Agar aap is lecture ko khud code karna chahte ho toh ye sequence follow karo:

### Step 1: Utility Package Banao (Random Data Generator)
Humein tests mein naye-naye random names aur numbers chahiye honge.

📁 **Folder banao:** `util`
📁 **File banao:** `util/random.go`
- Isme `init()` function me `rand.Seed` lagao.
- `RandomString`, `RandomInt`, `RandomOwner`, `RandomMoney`, aur `RandomCurrency` functions likho.

### Step 2: Test Entry Point Banao
Golang mein kisi ek package (`db`) ke saare tests run karne se pehle ek baar setup run karne ke liye `TestMain` use hota hai.

📁 **File banao:** `db/sqlc/main_test.go`
- Isme database connection (`sql.Open`) lagao.
- Global `testQueries` variable initialize karo.

### Step 3: Account ke CRUD Tests Likho
Ab actual test cases likho.

📁 **File banao:** `db/sqlc/account_test.go`
- Pehle ek helper function `createRandomAccount` banao jo baaki tests mein kaam aayega.
- Phir ek-ek karke: `TestCreateAccount`, `TestGetAccount`, `TestUpdateAccount`, `TestDeleteAccount`, aur `TestListAccounts` functions likho.

### Step 4: Makefile Update Karo
Terminal mein type karna aasan banane ke liye Makefile mein ek shortcut add karo.

📁 **Edit karo:** `Makefile`
```makefile
test:
	go test -v -cover ./...
```

### Step 5: Test Run Karo
Terminal mein ye chalao:
```bash
make test
```
Saare tests green (PASS) hone chahiye! ✅

---

## 🛠 Code Explanation 1: Testing Setup (`main_test.go`)

Har test ko run karne ke liye hume database connection ki zaroorat padegi. Bar-bar connection open/close karne se test slow ho jayenge. Isliye hum `TestMain` banate hain jo test suite start hone se pehle ek hi baar connection open karta hai.

```go
package db

import (
	"database/sql"
	"log"
	"os"
	"testing"
	_ "github.com/lib/pq" // Postgres driver zaroori hai DB se baat karne ke liye
)

// Global variable jisko saare tests (account_test.go wagerah) use karenge
var testQueries *Queries 

// TestMain main entry point hai
func TestMain(m *testing.M) {
	// Database connection open kiya
	conn, err := sql.Open("postgres", "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable")
	if err != nil {
		log.Fatal("cannot connect to db:", err) // Error aayi toh yahin test rok do
	}

	// Jo conn bana, usko humare SQLC `New` function mein daal diya
	testQueries = New(conn)

	// m.Run() saare unit tests ko start karega aur status (pass/fail) os.Exit ko dega
	os.Exit(m.Run())
}
```

---

## 🎲 Code Explanation 2: Random Data Generator (`util/random.go`)

Tests me agar hum fixed values (jaise name="Apoorv") use karenge, toh doosri baar test run karne par `UNIQUE` constraint fail ho sakta hai. Isliye random data banana zaroori hai.

```go
func init() {
	// Seed ke bina Golang hamesha same sequence me random number dega!
	// Isliye current time use karte hain taaki har bar sequence naya ho.
	rand.Seed(time.Now().UnixNano()) 
}

// RandomInt ek random integer deta hai min aur max ke beech
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString string builder ka use karke ek random alphabet ka word banata hai
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}
```
*(Yahin se `RandomOwner`, `RandomMoney` call hote hain)*

---

## ✅ Code Explanation 3: Writing Tests (`account_test.go`)

### 3.1: Create Account Test ➕
```go
// createRandomAccount helper hai jo account database me push karta hai
func createRandomAccount(t *testing.T) Account {
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(), // Yahan Random data use kiya
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	// testQueries (jo main_test me banaya tha) use karke create kiya
	account, err := testQueries.CreateAccount(context.Background(), arg)
	
	// testify/require ki power! Ye automatic check kar dega error nil hai ya nahi.
	require.NoError(t, err) 
	require.NotEmpty(t, account)

	// Check karo jo input tha wahi DB mein save hua hai ya nahi
	require.Equal(t, arg.Owner, account.Owner)
	// etc...

	return account
}

// Ye main test function hai Golang ke liye (Naam "Test" se shuru hona compulsory hai)
func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}
```

### 3.2: Get Account Test 🔍
```go
func TestGetAccount(t *testing.T) {
	// Pehle ek account create karo taaki hum usko dhundh sakein
	account1 := createRandomAccount(t)
	
	// Ab uski ID use karke usko wapis fetch karo
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	
	// Dono ekdam same hone chahiye
	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	
	// Timestamp (samay) exactly compare nahi hote fraction ki wajah se
	// Isliye 'WithinDuration' use karte hain ki dono 1 second ke gap ke andar-andar ho.
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
}
```

### 3.3: Delete Account Test 🗑️
```go
func TestDeleteAccount(t *testing.T) {
	// Ek account banao
	account1 := createRandomAccount(t)
	
	// Usko Delete karo
	err := testQueries.DeleteAccount(context.Background(), account1.ID)
	require.NoError(t, err) // Delete successful raha?

	// Ab agar us account ko Get karne ki koshish karein toh..
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	
	// Error aani chahiye!
	require.Error(t, err)
	
	// Error "sql.ErrNoRows" (Record doesn't exist) aani chahiye
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, account2)
}
```
*(Aise hi `Update` aur `List` ke tests kaam karte hain!)*
