# 🏦 Simple Bank: The Ultimate Master Guide (Lectures 1–18)

> **Welcome to your all-in-one Master Guide!**  
> Yeh single document tumhare pooray project ka comprehensive roadmap hai. Chahe tum 2 mahine ke break ke baad aa rahe ho ya interview ke liye Golang aur backend concepts revise kar rahe ho — yeh document project ke architecture, features, Golang ke fundamentals, aur **every critical engineering tradeoff** ko crystal-clear layman + technical depth mein explain karta hai.

---

## 📑 Table of Contents
1. [The Big Picture in Plain English (Layman Overview)](#1-the-big-picture-in-plain-english-layman-overview)
2. [High-Level System Architecture](#2-high-level-system-architecture)
3. [Golang Core Concepts Masterclass (Learned via this Project)](#3-golang-core-concepts-masterclass-learned-via-this-project)
4. [Deep-Dive: All Features Built So Far (Lectures 1–18)](#4-deep-dive-all-features-built-so-far-lectures-118)
5. [The Engineering Tradeoffs & Architectural Decisions (The "WHY")](#5-the-engineering-tradeoffs--architectural-decisions-the-why)
6. [End-to-End Execution Flows (Life of a Request)](#6-end-to-end-execution-flows-life-of-a-request)
7. [Repository Map & File Directory](#7-repository-map--file-directory)
8. [Commands, Testing & Development Guide](#8-commands-testing--development-guide)
9. [Next Steps Roadmap (Lecture 19+)](#9-next-steps-roadmap-lecture-19)

---

## 1. The Big Picture in Plain English (Layman Overview)

### Simple Bank Kya Hai?
Imagine karo **Google Pay**, **PhonePe**, ya **PayPal** ka backend server. Tum ek real-world financial banking engine bana rahe ho jisme:
- Users register karte hain apna account banate hain.
- Unke paas multiple currencies (`USD`, `EUR`, `CAD`) ke accounts hote hain.
- Ek user dusre user ko securely paise bhej sakta hai (Money Transfer).
- System har transaction ka complete financial audit trail (`entries` aur `transfers` tables) maintain karta hai.

### Real-World Engineering Challenges Jo Humne Solve Kiye:
1. **Paise Hawa Mein Gayab Na Hon (ACID Transactions)**: Agar Account A se 500 nikal gaye par server crash ho gaya isse pehle ki Account B mein jama hon, toh kya hoga? Humne database transactions (`BEGIN` ... `COMMIT` / `ROLLBACK`) use kiye taaki ya toh poora transfer ho ya kuch bhi na badle.
2. **Double Spending Aur Deadlocks**: Agar User A aur User B dono ek hi second mein ek-dusre ko paise bhejein, toh traditional databases hang ho jaate hain (Deadlock: A B ke lock hone ka wait karega aur B A ke lock hone ka). Humne algorithmically deadlock-free query ordering implement ki.
3. **Password Security**: Agar database leak ho jaye toh koi bhi user ka password na jaan sake (`bcrypt` adaptive hashing with automatic salt).
4. **Hacker-Proof Authentication**: Stateless digital passport system banaya (JWT aur modern PASETO tokens) taaki server par har request ke liye DB query na karni pade, jabki forgery namumkin ho.

---

## 2. High-Level System Architecture

```
┌────────────────────────────────────────────────────────────────────────┐
│                        HTTP Client (Postman / Web App)                │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ JSON over HTTP
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│  API Layer (`proj/api/`)                                               │
│  - Gin Engine: URL Routing (`/users`, `/accounts`, `/transfers`)       │
│  - Custom Validator: Currency check (`USD`, `EUR`, `CAD`)              │
│  - Handlers: JSON binding, error formatting, HTTP status codes         │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ Method call on Store interface
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│  Service & Security Layers                                             │
│  - `proj/token/`: JWTMaker & PasetoMaker (Issue & Verify Access Tokens)│
│  - `proj/util/`: Bcrypt Password Hashing, Config via Viper             │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ DB Operations & SQL Queries
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│  Database Layer (`proj/db/sqlc/`)                                      │
│  - Store Interface: Combines single queries + Atomic Transactions      │
│  - Type-Safe Generated Code: Created by `sqlc` directly from raw SQL   │
│  - Atomic Balance Update: `AddAccountBalance` with row locking         │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ SQL Wire Protocol (lib/pq)
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│  PostgreSQL Database (Docker Container)                                │
│  - Tables: `users` <──(FK)── `accounts` <──(FK)── `entries`            │
│                                    └──(FK)── `transfers`               │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Golang Core Concepts Masterclass (Learned via this Project)

Agar tum Go pehle se jante ho ya bhool chuke ho, yeh sections project ke code ke context mein Go ke 7 foundational pillars ko clear karte hain:

### 1. Packages & Visibility (Public vs Private)
- Go mein code folders (`packages`) mein organize hota hai.
- **Rule of Thumb**: Agar function/struct/variable ka pehla letter **Capital** hai (`Server`, `CreateUser`, `NewJWTMaker`), toh woh package ke bahar **Exported (Public)** hai. Agar **lowercase** hai (`errorResponse`, `minSecretKeySize`), toh woh **Internal (Private)** hai.

### 2. Structs & Receiver Methods (No Classes in Go!)
Go mein `class` ya `inheritance` nahi hoti. Data ko store karne ke liye `struct` banta hai aur functions ko struct ke saath bind karne ke liye **Receiver Methods** bante hain:
```go
// Data structure
type Server struct {
    store  db.Store
    router *gin.Engine
}

// Receiver method: (server *Server) binds this function to the Server struct
func (server *Server) Start(address string) error {
    return server.router.Run(address)
}
```
> [!NOTE]
> **Pointer Receiver `(s *Server)` vs Value Receiver `(s Server)`**:  
> Hum `*Server` (pointer) use karte hain taaki method pooray struct ki nayi copy memory mein na banaye, balki original memory address par operate kare.

### 3. Interfaces (Contracts & Duck Typing)
Interface Go ka sabse powerful feature hai. Interface define karta hai ki *kya behavior expected hai*:
```go
type Maker interface {
    CreateToken(username string, duration time.Duration) (string, error)
    VerifyToken(token string) (*Payload, error)
}
```
Go mein explicit `implements` keyword nahi hota. Agar kisi struct (`JWTMaker` ya `PasetoMaker`) ke paas yeh dono methods hain, toh Go automatically usse `Maker` maan leta hai ("Duck Typing").

### 4. Explicit Error Handling (No try/catch!)
Go mein exceptions nahi hote. Functions result ke saath error return karte hain:
```go
user, err := server.store.CreateUser(ctx, arg)
if err != nil {
    ctx.JSON(http.StatusInternalServerError, errorResponse(err))
    return
}
```
Is pattern se error silently ignore nahi ho sakta; developer ko har failure branch ko explicitly handle karna padta hai.

### 5. Goroutines & Concurrency Safety
Go lightweight threads chalata hai jinko **goroutines** (`go func()`) bolte hain. Humne `store_test.go` mein 10 parallel goroutines chala kar test kiya ki simultaneous transfers hone par database deadlock ya race condition toh generate nahi karta.

### 6. Struct Tags (Metadata)
Struct fields ke aage backticks mein tags lagte hain jo runtime par libraries use karti hain:
```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,alphanumeric"`
    Password string `json:"password" binding:"required,min=6"`
}
```
- `json:"username"`: JSON serialization/deserialization ke liye key name.
- `binding:"required,min=6"`: Gin ke validator ke rules.

---

## 4. Deep-Dive: All Features Built So Far (Lectures 1–18)

### 🔹 Feature 1: Database Schema & Migrations (Lectures 1–3, 15)
- **Problem**: Raw SQL scripts manually run karna production mein dangerous aur unreproducible hota hai.
- **Solution**: [`golang-migrate`](https://github.com/golang-migrate/migrate) tool use kiya with paired `.up.sql` (schema create karne) aur `.down.sql` (schema undo karne) files:
  - [`000001_init_schema.up.sql`](file:///c:/Users/apoor/Desktop/golang/proj/db/migration/000001_init_schema.up.sql): `accounts`, `entries`, aur `transfers` tables banaye with BigSerial IDs, Timestamps, aur Foreign Keys.
  - [`000002_add_users.up.sql`](file:///c:/Users/apoor/Desktop/golang/proj/db/migration/000002_add_users.up.sql): `users` table add ki aur `accounts.owner` ko `users.username` se link kiya with composite unique constraint `(owner, currency)` (ek user ke paas ek currency ka ek hi account ho sakta hai).

### 🔹 Feature 2: Type-Safe CRUD with `sqlc` (Lectures 4–6)
- **Problem**: Manually SQL string likhna aur `rows.Scan(&id, &owner...)` karna runtime runtime errors aur typos create karta hai.
- **Solution**: Raw declarative SQL queries likhi ([`account.sql`](file:///c:/Users/apoor/Desktop/golang/proj/db/query/account.sql), [`user.sql`](file:///c:/Users/apoor/Desktop/golang/proj/db/query/user.sql)). `sqlc` compiler ne compile-time par compile karke pure Go boilerplate methods generate kiye ([`account.sql.go`](file:///c:/Users/apoor/Desktop/golang/proj/db/sqlc/account.sql.go)).

### 🔹 Feature 3: Money Transfer Database Transaction (Lectures 7–9)
- **Problem**: Money transfer ek simple update nahi hai. Isme 5 steps hote hain:
  1. Record a `transfers` row.
  2. Record sender's negative `entries` row.
  3. Record receiver's positive `entries` row.
  4. Deduct sender balance.
  5. Add receiver balance.
- **Solution**: [`Store`](file:///c:/Users/apoor/Desktop/golang/proj/db/sqlc/store.go) struct banaya jisme `ExecTx` helper function SQL `BEGIN`, query execution, aur `COMMIT`/`ROLLBACK` ko safely encapsulate karta hai.

### 🔹 Feature 4: Deadlock Prevention Algorithm (Lecture 8)
- **Problem**: Jab Account 1 Account 2 ko paise bhej raha ho, aur usi millisecond mein Account 2 Account 1 ko bhej raha ho:
  - Tx 1: Locks Account 1, wants lock on Account 2.
  - Tx 2: Locks Account 2, wants lock on Account 1.
  - **Result**: Fatal database deadlock error!
- **Solution**: **Consistent Locking Order**. Humne code mein check lagaya:
  ```go
  if arg.FromAccountID < arg.ToAccountID {
      result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
  } else {
      result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
  }
  ```
  Hamesha choti ID waale account ko pehle lock karo! Isse circular dependency impossible ban jaati hai.

### 🔹 Feature 5: REST API Engine with Gin (Lectures 11, 14, 15)
- **File**: [`proj/api/server.go`](file:///c:/Users/apoor/Desktop/golang/proj/api/server.go)
- Endpoints implement kiye:
  - `POST /users`: User registration with hashed password.
  - `POST /accounts`: Bank account create karna.
  - `GET /accounts/:id`: Account fetch karna by ID.
  - `GET /accounts`: Paginated accounts list karna (`page_id`, `page_size`).
  - `POST /transfers`: Do accounts ke beech money transfer trigger karna.

### 🔹 Feature 6: Environment Configuration via Viper (Lecture 12)
- **File**: [`proj/util/config.go`](file:///c:/Users/apoor/Desktop/golang/proj/util/config.go)
- Hardcoded database credentials aur port hatakar [`app.env`](file:///c:/Users/apoor/Desktop/golang/proj/app.env) file banayi. Viper automatic unmarshaling karta hai into `Config` struct.

### 🔹 Feature 7: Custom Gin Validator (Lecture 14)
- **File**: [`proj/api/validator.go`](file:///c:/Users/apoor/Desktop/golang/proj/api/validator.go)
- Custom validator register kiya `currency` tag ke liye. Sirf allowed currencies (`USD`, `EUR`, `CAD`) accept hoti hain; invalid currency aane par request handler tak pahunchne se pehle hi `400 Bad Request` reject ho jati hai.

### 🔹 Feature 8: Secure Password Hashing with Bcrypt (Lecture 16)
- **File**: [`proj/util/password.go`](file:///c:/Users/apoor/Desktop/golang/proj/util/password.go)
- Password ko plaintext mein store karne ke bajaye `bcrypt.GenerateFromPassword` (default cost 10) use karke secure salt + hash string banaya. Verify karne ke liye `bcrypt.CompareHashAndPassword` use kiya.

### 🔹 Feature 9: Unit Testing with Gomock & Custom Matchers (Lectures 13, 17)
- **Files**: [`proj/api/account_test.go`](file:///c:/Users/apoor/Desktop/golang/proj/api/account_test.go), [`proj/api/user_test.go`](file:///c:/Users/apoor/Desktop/golang/proj/api/user_test.go)
- Real database ke bina fast unit tests likhe. Gomock se mock store generate kiya, aur custom matcher banaya jo verify karta hai ki plain password aur hashed password match hote hain ya nahi without knowing the random salt.

### 🔹 Feature 10: Token Authentication (JWT & PASETO) (Lecture 18)
- **Files**: [`proj/token/maker.go`](file:///c:/Users/apoor/Desktop/golang/proj/token/maker.go), [`proj/token/jwt_maker.go`](file:///c:/Users/apoor/Desktop/golang/proj/token/jwt_maker.go), [`proj/token/paseto_maker.go`](file:///c:/Users/apoor/Desktop/golang/proj/token/paseto_maker.go)
- Generics-ready `Maker` interface banaya jisme token generation aur verification encapsulated hai.
- Dono JWT aur PASETO makers ke comprehensive unit tests likhe ([`jwt_maker_test.go`](file:///c:/Users/apoor/Desktop/golang/proj/token/jwt_maker_test.go), [`paseto_maker_test.go`](file:///c:/Users/apoor/Desktop/golang/proj/token/paseto_maker_test.go)).

---

## 5. The Engineering Tradeoffs & Architectural Decisions (The "WHY")

Har design decision ke piche engineering trade-offs hote hain. Interviewers aur senior engineers yeh sabse zyada puchte hain:

| Decision Area | What We Chose | Alternative Rejected | Why? The Trade-Off & Rationale |
| :--- | :--- | :--- | :--- |
| **Testing Persona** | `Store` Interface | Concrete `*SQLStore` struct | **Testing Isolation**: Agar `Server` concrete DB struct use karega, toh har unit test ke liye real Postgres container chahiye hoga (slow, flaky, shared state). Interface use karke hum Gomock se in-memory fake store inject kar sakte hain jo microsecond mein test execute karta hai. |
| **Database Query Tool** | `sqlc` | GORM / SQLX / Raw SQL | **Performance & Compile-Time Safety**: GORM bohot heavy reflection use karta hai aur complex joins/locks par unoptimized queries likhta hai. SQLX type safe nahi hai. `sqlc` raw SQL ki speed deta hai + compile time par syntax & type mismatch check kar leta hai. |
| **Database Locking** | Direct Arithmetic Query (`AddAccountBalance`) | `SELECT FOR UPDATE` + Read + Update | **Deadlock & Query Reduction**: Pehle `SELECT FOR UPDATE` karke Go mein `balance = balance - amount` karke update bhejte the. Lekin foreign key check par Postgres entries table par lock laga deta tha. `AddAccountBalance` (`SET balance = balance + $1`) single atomic operation hai jo foreign key lock ko bypass karta hai. |
| **Deadlock Avoidance** | Deterministic ID Ordering | Distributed Lock / Retry Queue | **Zero Overhead**: Retry queue ya Redis locks latency aur complexity badhate hain. Account IDs ko sort karke (`min(A, B)` pehle, `max(A, B)` baad mein) circular lock dependency mathematically impossible ho jaati hai bina kisi extra service ke. |
| **Password Storage** | `bcrypt` (Adaptive Hashing) | MD5 / SHA256 / PBKDF2 | **Brute-Force Resistance**: MD5 aur SHA256 bohot fast hote hain — ek high-end GPU billions of hashes per second compute kar sakti hai. `bcrypt` intentionally slow hota hai (work factor/cost) aur automatically random salt add karta hai to defeat rainbow tables. |
| **Auth Tokens** | `PASETO` (over JWT) | JSON Web Tokens (JWT) | **Security by Default**: JWT standard bohot prone to attacks hai: algorithm header (`alg: none`) exploit, RSA vs HMAC algorithm confusion. PASETO mein koi algorithm header nahi hota; modern cryptography (ChaCha20-Poly1305) hardcoded hoti hai. |
| **Token Encryption** | Symmetric (`Local` / Shared Key) | Asymmetric (`Public/Private` Keypair) | **Simplicity & Latency**: Kyunki abhi humara monolithic backend hi token issue aur verify kar raha hai, 32-byte shared symmetric key faster aur simpler hai. Agar future mein dedicated Auth microservice banta hai, hum asymmetric par switch kar sakte hain bina handlers ko change kiye (thanks to `Maker` interface). |
| **Config Loader** | `Viper` | Standard `os.Getenv()` | **Type Safety & Defaults**: `os.Getenv` se string parsing manually karni padti hai. Viper unmarshals into a typed struct, supports `.env`, environment variables, default values, aur test overrides seamlessly. |
| **HTTP Framework** | `Gin` | Standard `net/http` | **Developer Velocity**: Go ka `net/http` bohot low-level hai (no path parameter extraction `/accounts/:id`, no built-in JSON binding). Gin provide karta hai fast Radix-tree router aur validator integration with ultra-low memory footprint. |

---

## 6. End-to-End Execution Flows (Life of a Request)

### Flow A: User Registration (`POST /users`)
```
1. Client sends POST request with { "username": "alice", "password": "secretpassword", "full_name": "Alice Smith", "email": "alice@gmail.com" }
2. Gin Router matches route -> invokes api.server.createUser handler
3. Validator checks: username alphanumeric, password length >= 6, email valid format
4. Util Layer: HashPassword("secretpassword") calls bcrypt -> produces 60-character hash
5. Store Layer: store.CreateUser(ctx, CreateUserParams{ ... HashedPassword: hash })
6. Database: INSERT INTO users ... RETURNING *
7. Handler sanitizes output: copies fields into userResponse (excludes HashedPassword)
8. Client receives 200 OK with sanitized JSON profile
```

### Flow B: Money Transfer (`POST /transfers`)
```
1. Client sends POST request: { "from_account_id": 1, "to_account_id": 2, "amount": 100, "currency": "USD" }
2. Gin Validator checks: positive amount, valid currency, from_account_id != to_account_id
3. Handler checks: both accounts exist and currencies match request currency
4. Store.TransferTx starts:
   a. tx.CreateTransfer: saves transfer audit record
   b. tx.CreateEntry: records -100 for Account 1
   c. tx.CreateEntry: records +100 for Account 2
   d. Compare IDs: 1 < 2 -> update Account 1 (-100) first, then Account 2 (+100)
   e. SQL COMMIT executes atomically
5. Handler receives TransferTxResult and returns HTTP 200 OK
```

---

## 7. Repository Map & File Directory

```
golang/
├── PROJECT_OVERVIEW.md                  # This root Master Guide
├── go.work                              # Go workspace definition
└── proj/
    ├── PROJECT_OVERVIEW.md              # In-project copy of Master Guide
    ├── Makefile                         # Automation commands (test, migrate, server)
    ├── app.env                          # Local environment variables & secrets
    ├── main.go                          # Application entrypoint
    ├── api/                             # HTTP Transport Layer
    │   ├── server.go                    # Gin Router setup & route registration
    │   ├── account.go                   # Handlers for Account endpoints
    │   ├── account_test.go              # Mock unit tests for Account API
    │   ├── user.go                      # Handlers for User endpoints
    │   ├── user_test.go                 # Mock unit tests for User API (custom gomock)
    │   ├── transfer.go                  # Handlers for Money Transfer
    │   ├── validator.go                 # Custom ISO currency validation
    │   └── main_test.go                 # Gin test setup (ReleaseMode)
    ├── db/                              # Database Layer
    │   ├── migration/                   # Schema migration scripts (000001, 000002)
    │   ├── query/                       # Raw SQL queries (account.sql, user.sql, etc.)
    │   ├── mock/                        # Gomock generated Store mock for API unit tests
    │   └── sqlc/                        # Generated type-safe Go database code
    │       ├── store.go                 # ACID Database Transaction logic (TransferTx)
    │       ├── store_test.go            # Concurrency & deadlock unit tests
    │       └── ...                      # Generated models & queries
    ├── token/                           # Authentication & Security
    │   ├── maker.go                     # Token Maker interface & Payload contract
    │   ├── payload.go                   # Token payload struct & UUID generation
    │   ├── jwt_maker.go                 # JWT implementation with HMAC-SHA256
    │   ├── jwt_maker_test.go            # JWT tests (expiry, none-algorithm attack)
    │   ├── paseto_maker.go              # PASETO v2 implementation (ChaCha20-Poly1305)
    │   └── paseto_maker_test.go         # PASETO tests (creation, expiry)
    └── util/                            # Helpers & Utilities
        ├── config.go                    # Viper config loading
        ├── password.go                  # Bcrypt hashing & verification
        ├── password_test.go             # Password hashing unit tests
        ├── currency.go                  # Currency constants (USD, EUR, CAD)
        └── random.go                    # Fake test data generators
```

---

## 8. Commands, Testing & Development Guide

Hamesha `proj/` directory ke andar commands run karein:
```powershell
cd c:\Users\apoor\Desktop\golang\proj
```

### 1. Run Unit Tests (No Database Needed! ⚡)
Yeh tests Gomock aur in-memory mocks use karte hain, isliye Docker/Postgres ke bina instant run hote hain:
```powershell
# Run all fast unit tests (API, Token Makers, Passwords)
go test -v ./api/... ./token/... ./util/...
```

### 2. Run Database Integration Tests (Requires Postgres Container)
```powershell
# Start docker container
make postgres

# Run DB schema migrations
make migrateup

# Run all tests including DB transactions and deadlock tests
go test -v -race ./db/sqlc/...
```

### 3. Start Local Server
```powershell
go run main.go
# Server starts on http://localhost:8080
```

### 4. Regenerate Code (Mocks & Sqlc)
```powershell
# Regenerate Gomock mock store
make mock

# Regenerate SQLC Go code from query/*.sql
make sqlc
```

---

## 9. Next Steps Roadmap (Lecture 19+)

Tumhara code currently **Lecture 18** tak complete, fully tested, aur git repo mein committed hai!

Aage ka roadmap:
1. **Lecture 19: User Login API (`POST /users/login`)**:
   - Request: `{ "username": "alice", "password": "secretpassword" }`
   - Steps: Database se user fetch karna -> `CheckPassword` verify karna -> `PasetoMaker.CreateToken` se access token generate karna.
   - Response: Return `{ "access_token": "...", "user": { ... } }`.
2. **Lecture 20: Authentication Middleware**:
   - Gin middleware likhna jo `Authorization: Bearer <token>` header read karega.
   - `maker.VerifyToken` call karke context (`ctx.Set("authorization_payload", payload)`) mein store karega.
3. **Lecture 21: Route Protection**:
   - Transfer API aur Account creation API ko protect karna taaki logged-in user sirf apne account se hi transfer initiate kar sake.

---
*Ready to build? Whenever you want to start Lecture 19, just say the word!*
