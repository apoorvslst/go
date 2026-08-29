# 💸 Lecture 14: Transfer Money API & Custom Validator in Go

## 🎯 High Level Overview

Is lecture mein humne apni application ka sabse important feature banaya hai: **Transfer Money API**. Iske saath hi humne ek bohot important concept seekha hai: **Custom Validator**.

### Transfer API ka Flow
Jab koi user paisa transfer karta hai (e.g. Account A se Account B mein 10 USD), toh server ko kuch cheezein check karni hoti hain:
1. Kya Account A exist karta hai?
2. Kya Account B exist karta hai?
3. Kya dono accounts ki currency match karti hai? (Hum USD account se EUR account mein direct transfer support nahi kar rahe).
4. Agar sab theek hai, toh `TransferTx` (Database Transaction) execute karo jo humne pichle lectures mein banayi thi.

### Custom Validator Kyu Chahiye?
Gin framework mein parameters check karne ke liye hum tags use karte hain, jaise `binding:"required,oneof=USD EUR CAD"`. 
**Problem:** Agar kal ko hum 100 currencies support karte hain, toh har jagah (CreateAccount API, Transfer API) 100 currencies ka bada string likhna padega. Ye messy aur hard to maintain hai.
**Solution:** Hum apna khud ka custom validation tag banayenge, jaise `binding:"currency"`, jo automatically check karega ki input currency supported hai ya nahi.

---

## 📝 Correct Order of Writing Code (Agar Khud Se Karna Ho)

Agar aap isko khud implement kar rahe hain, toh ye steps follow karein:

### Phase 1: Basic Transfer API setup
1. **`api/transfer.go`** — Nayi file banao, `transferRequest` struct likho.
2. **`api/transfer.go`** — `createTransfer` handler function likho (basic structure).
3. **`api/transfer.go`** — `validAccount` helper function likho jo DB se account check kare aur currency verify kare.
4. **`api/server.go`** — Nayi route register karo: `router.POST("/transfers", server.createTransfer)`.

### Phase 2: Custom Validator Implementation
5. **`util/currency.go`** — Nayi file banao. Saari supported currencies ko as constants define karo aur `IsSupportedCurrency` function likho.
6. **`api/validator.go`** — Nayi file banao. `validCurrency` function likho jo `validator.FieldLevel` accept kare.
7. **`api/server.go`** — `NewServer` function mein `binding.Validator.Engine()` use karke custom validator register karo.
8. **`api/account.go` & `api/transfer.go`** — `oneof=USD EUR` hata kar apna custom tag `currency` use karo.
9. **`util/random.go`** — `RandomCurrency` function ko update karo taaki wo `util/currency.go` ke constants use kare.

---

## 🛠️ Step 1: `api/transfer.go` (Input Struct & Basic Handler)

Nayi file mein sabse pehle hum input data ka structure define karte hain.

```go
package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/apoor/simple_bank/db/sqlc"
	"github.com/gin-gonic/gin"
)

// transferRequest API request ka format define karta hai
type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"` // Custom validator tag!
}
```
**Explanation:**
- `FromAccountID` aur `ToAccountID`: Paise kahan se kahan jayenge. `min=1` kyu? Kyunki DB IDs 1 se start hoti hain.
- `Amount`: Kitna paisa. `gt=0` (greater than 0) use kiya taaki koi negative ya zero amount transfer na kar sake.
- `Currency`: Kis currency mein transfer ho raha hai. Yahan hum baad mein apna custom `currency` tag lagayenge.

```go
func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	
	// Step 1: Request check karo
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Step 2: Validations check karo (Account exist karte hain? Currency theek hai?)
	if !server.validAccount(ctx, req.FromAccountID, req.Currency) {
		return
	}
	if !server.validAccount(ctx, req.ToAccountID, req.Currency) {
		return
	}

	// Step 3: Transaction execute karo
	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Step 4: Success! Response return karo
	ctx.JSON(http.StatusOK, result)
}
```
**Explanation:**
- Yeh handler pehle user ka JSON data read karta hai. 
- Uske baad ek helper function `validAccount` use karke check karta hai ki kya fromAccount aur toAccount exist karte hain, aur kya unki currency match karti hai.
- Agar validations theek hain, toh DB transaction `TransferTx` call hoti hai.

---

## 🛠️ Step 2: `api/transfer.go` (The `validAccount` Helper)

Ye function check karta hai ki kya kisi particular account ki ID valid hai aur uski currency input currency se match karti hai ya nahi.

```go
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) bool {
	// DB se account nikalo
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Account mila hi nahi
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		// DB mein kuch internal error ho gaya
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	// Currency match karni chahiye
	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	return true // Account bilkul theek hai!
}
```
**Explanation:**
- `GetAccount`: Pehle database mein check karte hain account hai ya nahi.
- `sql.ErrNoRows`: Agar nahi milta, toh 404 (Not Found) return karke `false` bhej dete hain, jisse parent function vahin ruk jata hai.
- `account.Currency != currency`: Custom error banate hain taaki client ko properly samajh aaye ki usne galat currency bheji hai (e.g., account CAD ka hai par USD bhej diya). 400 (Bad Request) status dete hain.

---

## 🛠️ Step 3: `api/server.go` (Register Transfer Route)

Naya API endpoint server mein add karna padega.

```go
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// Existing routes...
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)

	// Nayi route add ki!
	router.POST("/transfers", server.createTransfer)

	server.router = router
	return server
}
```

---

## 🛠️ Step 4: `util/currency.go` (Custom Validator ke liye Constants)

Ab hum custom validator ki taraf badhte hain. Pehle saari supported currencies ko ek common jagah par likhte hain.

```go
package util

// Constants for all supported currencies
const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

// IsSupportedCurrency returns true if the currency is supported
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, CAD:
		return true
	}
	return false
}
```
**Explanation:**
- Hardcoded strings (jaise `"USD"`) ko constant banaya. Ye best practice hai kyunki spelling mistakes catch ho jati hain (compiler pakad leta hai).
- Ek helper function `IsSupportedCurrency` banaya jo switch statement use karke check karta hai. Agar kal nayi currency add karni hui (e.g. INR), toh bas is file mein update karna padega!

---

## 🛠️ Step 5: `api/validator.go` (Custom Validator ka Logic)

Gin underlying package `go-playground/validator/v10` use karta hai. Hum apna khud ka validation function banayenge.

```go
package api

import (
	"github.com/apoor/simple_bank/util"
	"github.com/go-playground/validator/v10"
)

// validCurrency is a custom validator function for Gin
var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	// 1. Field ki value nikalo aur String mein convert karo
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		// 2. Check karo ki kya string humari supported list mein aati hai
		return util.IsSupportedCurrency(currency)
	}
	// Agar wo string hi nahi hai, toh false return karo
	return false
}
```
**Explanation:**
- `validator.Func`: Ye ek specific type ka function hai jo Gin ka validator expect karta hai. Iska input `validator.FieldLevel` hota hai.
- `fieldLevel.Field().Interface().(string)`: Reflection (`reflect` package concept) use karke value nikal rahe hain aur type assertion se usko `string` mein convert kar rahe hain.
- Agar wo sach mein string hai, toh apne `util` wale function mein bhej ke true/false return kar dete hain.

> [!TIP]
> **🟢 Node.js Equivalent: Custom Validator (e.g. express-validator / Joi)**
> Node mein hum Joi ya Zod jaisi libraries me directly array include match laga sakte hain ya custom validation message likh sakte hain. Gin me ye logic slightly low-level hota hai reflection ke karan.

---

## 🛠️ Step 6: `api/server.go` (Registering Custom Validator with Gin)

Function toh bana liya, par Gin ko kaise pata chalega ki is function ko kis naam (`tag`) se bulana hai?

```go
import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// ⭐️ Custom Validator Register karne ka code ⭐️
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// "currency" tag ka naam hai jo struct me use hoga
		v.RegisterValidation("currency", validCurrency)
	}

	router.POST("/accounts", server.createAccount)
	// ... (baaki routes)
```
**Explanation:**
- `binding.Validator.Engine()`: Gin ka internal validation engine nikalte hain.
- `.(*validator.Validate)`: Type casting karte hain version 10 wale validator pointer mein.
- `v.RegisterValidation("currency", validCurrency)`: Gin ko bolte hain ki "Bhai, jab bhi struct me `binding:"currency"` tag mile, toh `validCurrency` wala function run kar dena".

---

## 🛠️ Step 7: Structs ko Update Karna

Ab hum apna naya custom tag use kar sakte hain!

**`api/account.go` mein:**
```go
type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	// PEHLE: Currency string `json:"currency" binding:"required,oneof=USD EUR"`
	// AB:
	Currency string `json:"currency" binding:"required,currency"` 
}
```

**`api/transfer.go` mein:**
```go
type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}
```

**`util/random.go` mein:**
Hum strings ko constants se replace kar dete hain taaki consistency rahe.
```go
func RandomCurrency() string {
	currencies := []string{USD, EUR, CAD}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}
```

---

## 🚀 Summary (Faayda Kya Hua?)
- **DRY Principle (Don't Repeat Yourself)**: Pehle `oneof=USD EUR CAD` har us struct mein likhna padta jahan currency input hai. Agar ek chhut gaya toh bug aa jayega. Ab sirf `currency` likhna hai!
- **Easy Updates**: Agar future mein `INR` add karna hai, toh humein API ke structs ko touch nahi karna padega. Sirf `util/currency.go` mein jake constant add kardo. 
- **Robustness**: Transfer logic mein humne explicitly check kiya ki Account A aur Account B same currency share karein, otherwise reject kar do. Ye ek complete real-world transaction flow ban gaya hai.
