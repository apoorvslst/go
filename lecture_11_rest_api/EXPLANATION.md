# 🔄 Lecture 11: RESTful HTTP API with Gin Framework (Detailed & Easy Explanation)

Is lecture mein humne apna pehla **RESTful HTTP API** server banaya hai. Agar aap bilkul beginner hain, toh is guide ko padhiye, har ek line of code ka matlab simple words mein samjhaya gaya hai.

---

## 🧐 RESTful API aur Gin Kya Hota Hai?

Jab koi user aapki app (frontend) kholta hai, toh app ko data kahan se milta hai? **Backend se!**
Frontend aur Backend aapas mein **API** (Application Programming Interface) ke through baat karte hain. REST API bas ek standard tareeqa hai jisme URLs (`/accounts`) aur methods (`GET`, `POST`) ka use hota hai.

Golang mein API banane ke liye bahut saare tools hain. Hum **Gin Framework** use kar rahe hain kyunki:
- Yeh bahut fast hai.
- Yeh apne aap check kar leta hai ki user ne sahi data bheja hai ya nahi (validation).
- Isme code likhna aasan hota hai.

> [!TIP]
> **🔄 Node.js / Express.js Se Comparison:**
> Niche har Go snippet ke baad ek **🟢 Node.js Equivalent** block diya gaya hai. Agar aapko Express.js aata hai, toh Gin framework samajhna aapke liye bohot aasan hoga, dono ka pattern lagbhag same hai!

---

## 🛠️ File 1: `api/server.go` (Server ka Engine)

Is file mein hum apna server setup karte hain aur usko batate hain ki kaunsa URL aane par kya kaam karna hai.

```go
package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/apoor/simple_bank/db/sqlc"
)

// Server struct ek aisi object hai jo hamare server ka saara state hold karegi
type Server struct {
	store  *db.Store       // Database se baat karne ke liye connection (pichle lectures ka)
	router *gin.Engine     // Gin ka router jo URLs ko sahi function ke paas bhejta hai
}
```
**Simple Bhasha Mein:** `Server` ek dibba (struct) hai jiske andar do cheezein hain:
1. `store`: Jiske paas Database ka access hai (database se records lana/dalna).
2. `router`: Ek traffic police 👮 jo dekhta hai ki user kis URL par aaya hai aur usko kis function ke paas bhejna hai.

```go
// NewServer ek factory function hai jo naya Server object banata hai
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default() // Gin ka default router banaya

	// 🚦 Routing Setup (Traffic Police ka rules)
	router.POST("/accounts", server.createAccount)   // Naya account banaye
	router.GET("/accounts/:id", server.getAccount)   // ID ke hisaab se ek account laaye
	router.GET("/accounts", server.listAccount)      // Saare accounts ki list laaye

	server.router = router // Router ko Server ke andar save kar diya
	return server
}
```
**Routing ka logic:**
- Jab koi `POST` request bhejega `/accounts` par, toh `createAccount` naam ka function chalega.
- Jab koi `GET` request bhejega `/accounts/1` par, toh `getAccount` function chalega. (`:id` ka matlab hai yahan koi bhi number aa sakta hai).

> [!TIP]
> **🟢 Node.js (Express) Equivalent: App Setup & Routing**
> ```javascript
> const express = require('express');
> const app = express();
> app.use(express.json()); // Gin ka gin.Default() jaisa
> 
> app.post('/accounts', createAccount);
> app.get('/accounts/:id', getAccount);
> app.get('/accounts', listAccount);
> ```

```go
// Start function server ko chalu karta hai aur requests ka intezar karta hai
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// Ye ek chhota helper function hai jo simple text error ko JSON format mein badalta hai
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
```
**JSON kyu?** Kyunki Frontend ko JSON format samajh aata hai. `gin.H` bas ek shortcut hai `map[string]interface{}` ka, jo automatically `{ "error": "kuch gadbad hai" }` ban jata hai.

> [!TIP]
> **🟢 Node.js (Express) Equivalent: Start Server & Error Response**
> ```javascript
> app.listen(8080, () => console.log("Server running on port 8080"));
> 
> function errorResponse(err) {
>   return { error: err.message };
> }
> ```

---

## 🛠️ File 2: `api/account.go` (Asli Kaam Karne Wale Functions)

### 1. Naya Account Banana (Create Account)

```go
// User se kya data chahiye uski list:
type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR"`
}
```
**Validation Tags (Jadoo 🪄):** 
- `binding:"required"` ka matlab hai ki user ko Owner name dena hi padega, warna Gin khud error de dega.
- `oneof=USD EUR` ka matlab hai currency ya toh "USD" honi chahiye ya "EUR". Agar kisi ne "INR" bhej diya toh Gin request reject kar dega.

```go
func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	
	// Step 1: User ke JSON data ko humare struct mein fit karo
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // 400 Bad Request
		return
	}

	// Step 2: Database mein bhejne ke liye data taiyar karo
	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0, // Naya account hamesha zero balance se start hota hai
	}

	// Step 3: Database mein save karo!
	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // 500 Server Error
		return
	}

	// Step 4: Success! Naya bana hua account data frontend ko wapas bhej do
	ctx.JSON(http.StatusOK, account)
}
```

> [!TIP]
> **🟢 Node.js (Express) Equivalent: Create Account**
> ```javascript
> async function createAccount(req, res) {
>   try {
>     const { owner, currency } = req.body;
>     if (!owner || !['USD', 'EUR'].includes(currency)) {
>        return res.status(400).json({ error: "Invalid data" });
>     }
>     
>     // DB Save Logic (like Sequelize/Mongoose)
>     const account = await db.Account.create({ owner, currency, balance: 0 });
>     
>     res.status(200).json(account);
>   } catch (err) {
>     res.status(500).json({ error: err.message });
>   }
> }
> ```

---

### 2. Kisi Ek Account Ki Details Lena (Get Account)

```go
type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}
```
**Fark samjho:** Yahan `uri:"id"` hai. Iska matlab ID JSON body se nahi aayegi, balki URL se nikalni hai (jaise `/accounts/5` me se `5` nikalna). `min=1` kyu? Kyunki account ID negative nahi ho sakti.

```go
func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	
	// ShouldBindUri URL se ':id' nikal kar req struct mein dalega
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Database se account mangwao
	account, err := server.store.GetAccount(ctx, req.ID)
	
	if err != nil {
		// Agar account exist hi nahi karta!
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err)) // 404 Not Found bhej do
			return
		}
		// Koi aur badi error ho gayi DB mein
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}
```

> [!TIP]
> **🟢 Node.js (Express) Equivalent: Get Account**
> ```javascript
> async function getAccount(req, res) {
>   try {
>     const id = parseInt(req.params.id);
>     if (isNaN(id) || id < 1) return res.status(400).json({ error: "Invalid ID" });
> 
>     const account = await db.Account.findByPk(id);
>     if (!account) return res.status(404).json({ error: "Not Found" });
> 
>     res.status(200).json(account);
>   } catch (err) {
>     res.status(500).json({ error: err.message });
>   }
> }
> ```

---

### 3. Ek Sath Bahut Saare Accounts Dekhna (List Accounts + Pagination)

Agar bank mein 10 lakh accounts hain, toh hum ek saath 10 lakh nahi bhej sakte. Hum usko "Pages" mein bhejte hain (e.g. Page 1 par 5 accounts). Isko **Pagination** kehte hain.

```go
type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}
```
**`form` tag kyu?** Kyunki yeh data URL query string mein hota hai: `/accounts?page_id=1&page_size=5`. `max=10` matlab client ek baar mein 10 se zyada account nahi mangwa sakta.

```go
func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	
	// ShouldBindQuery URL ke aage wale ?page_id=... ko read karta hai
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Math magic: Database ko kitne records chhodne (Skip/Offset) hain?
	// Agar Page 1 hai: (1-1)*5 = 0 records chhodne hain.
	// Agar Page 2 hai: (2-1)*5 = 5 records chhodne hain.
	arg := db.ListAccountsParams{
		Limit:  req.PageSize,                         // Kitne records laane hain
		Offset: (req.PageID - 1) * req.PageSize,      // Kitne records skip karne hain
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}
```

> [!TIP]
> **🟢 Node.js (Express) Equivalent: List Accounts + Pagination**
> ```javascript
> async function listAccount(req, res) {
>   try {
>     const pageId = parseInt(req.query.page_id);
>     const pageSize = parseInt(req.query.page_size);
> 
>     const limit = pageSize;
>     const offset = (pageId - 1) * pageSize;
> 
>     const accounts = await db.Account.findAll({ limit, offset });
>     res.status(200).json(accounts);
>   } catch (err) {
>     res.status(500).json({ error: err.message });
>   }
> }
> ```

---

## 🛠️ File 3: `main.go` (Start Button)

Sab kuch bana liya, ab in saare tukdon ko jod kar server on karna hai.

```go
func main() {
	// 1. Database ka darwaza (connection) kholo
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err) // Error aaye toh program band kardo
	}

	// 2. Apne Database queries (Store) ko is darwaze se jod do
	store := db.NewStore(conn)
	
	// 3. Apna API Server banao jisme yeh DB queries use hongi
	server := api.NewServer(store)

	// 4. Server ko port 8080 par On kardo (Listening mode)
	err = server.Start("0.0.0.0:8080")
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
```

> [!TIP]
> **🟢 Node.js (Express) Equivalent: Main / Entry Point**
> ```javascript
> // DB connection
> mongoose.connect('mongodb://localhost/simple_bank')
>   .then(() => {
>     console.log('Connected to DB');
>     app.listen(8080, () => console.log('Server started on port 8080'));
>   })
>   .catch(err => console.error('DB connection error:', err));
> ```

---

## 💡 Ek Akhiri "Jugaad" (sqlc.yaml wala empty array issue)
Default roop se, agar DB me koi data nahi hota tha toh Go list API return karti thi `null`. Lekin frontend walo ko `null` dekh kar gussa aata hai, unhe empty list `[]` chahiye hoti hai.
Toh humne `sqlc.yaml` mein jake bataya:
```yaml
emit_empty_slices: true
```
Toh ab `sqlc` ne DB code aise likha ki "Agar data nahi mila, toh `null` mat bhejo, khali dibba `[]` bhejo". Phir humne `sqlc generate` run kar diya. 

---

**Summary:** 
1. Client (Postman/Web) Request Bhejta hai.
2. Gin ka Router usko sahi Handler (`createAccount`, `getAccount`) tak pahuchata hai.
3. Handler user ke bheje data ko check karta hai (`ShouldBind...`).
4. Agar data theek hai, toh `Store` ke through Database se baat karta hai.
5. Result (ya Error) wapas JSON me Client ko bhej deta hai!
