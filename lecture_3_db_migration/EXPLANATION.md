# 🔄 Lecture 3: Database Migration in Golang

## Ye Lecture Kya Hai?

Is lecture mein sikhaya gaya hai ki **database schema migration** kya hota hai aur **golang-migrate** library kaise use karte hain. Plus kuch naye Docker commands bhi dikhaye gaye.

---

## 🤔 Migration Kya Hai? (Bahut Simple Example)

Socho tumne ek app banayi — "Simple Bank". Shuru mein tumne 3 tables banaye:
- accounts
- entries 
- transfers

Ab 2 mahine baad boss bolta hai: **"Har account mein email bhi chahiye!"**

Toh kya karoge? Directly database mein jaake `ALTER TABLE` chaloge? **GALAT! ❌**

Kyun? Kyunki:
1. Tumhare team ke doosre log ko pata nahi chalega ki tumne kya change kiya
2. Production mein same change karna bhool jaoge
3. Agar kuch galat ho jaaye toh **undo** (wapas) kaise karoge?

**Solution: Migration! ✅**

Migration = **Database ka version control** (jaise Git code ke liye hai)

```
Socho aise:

Git:       v1 → v2 → v3 → v4   (code ke versions)
Migration: v1 → v2 → v3 → v4   (database ke versions)

Git:       git revert (code wapas)
Migration: migrate down (database wapas)
```

---

## 📁 Migration Files Kaise Dikhte Hain?

Har migration mein **DO files** hoti hain:

```
db/migration/
├── 000001_init_schema.up.sql      ← AAGE jaao (tables banao)
├── 000001_init_schema.down.sql    ← PEECHHE jaao (tables hatao)
├── 000002_add_email.up.sql        ← AAGE: email column add karo
├── 000002_add_email.down.sql      ← PEECHHE: email column hatao
├── 000003_add_phone.up.sql        ← AAGE: phone column add karo
└── 000003_add_phone.down.sql      ← PEECHHE: phone column hatao
```

**UP = Aage jaao (change karo)**
**DOWN = Peechhe jaao (change undo karo)**

```
migrate up           migrate down
    │                    │
    ▼                    ▼
┌────────┐          ┌────────┐
│ v1: UP │──────────│ v3: DN │
│ tables │          │ phone  │
│ banao  │          │ hatao  │
├────────┤          ├────────┤
│ v2: UP │──────────│ v2: DN │
│ email  │          │ email  │
│ add    │          │ hatao  │
├────────┤          ├────────┤
│ v3: UP │──────────│ v1: DN │
│ phone  │          │ tables │
│ add    │          │ hatao  │
└────────┘          └────────┘

UP: 1 → 2 → 3 (order mein)
DOWN: 3 → 2 → 1 (ulta order mein)
```

---

## 🛠️ golang-migrate Install Karna

### Windows pe:

```bash
# Option 1: Scoop se (agar scoop hai)
scoop install migrate

# Option 2: Go install se
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Option 3: GitHub releases se download karo
# https://github.com/golang-migrate/migrate/releases
# Windows ka .exe download karo, PATH mein daal do
```

### Mac pe:
```bash
brew install golang-migrate
```

### Check karo install hua ya nahi:
```bash
migrate -version
# Output: 4.x.x
```

---

## 📝 Migration Files Banana

### Command:

```bash
migrate create -ext sql -dir db/migration -seq init_schema
```

**Har part ka matlab:**

| Part | Matlab |
|------|--------|
| `migrate create` | Naye migration files banao |
| `-ext sql` | File extension `.sql` rakho |
| `-dir db/migration` | Files yahan save karo |
| `-seq` | Sequential number do (000001, 000002...) |
| `init_schema` | Migration ka naam |

**Ye command 2 files banata hai:**
```
db/migration/000001_init_schema.up.sql    ← Isme tables banana likhenge
db/migration/000001_init_schema.down.sql  ← Isme tables delete karna likhenge
```

---

## 📄 Migration Files ka Code

### UP File — `000001_init_schema.up.sql`

Ye file chalegi jab `migrate up` karoge. Tables ban jayengi:

```sql
-- Accounts table: Bank accounts store karta hai
CREATE TABLE "accounts" (
  "id" bigserial PRIMARY KEY,          -- auto-increment ID
  "owner" varchar NOT NULL,            -- account owner ka naam
  "balance" bigint NOT NULL,           -- kitne paise hain
  "currency" varchar NOT NULL,         -- USD, INR, EUR...
  "created_at" timestamptz NOT NULL DEFAULT (now())  -- kab bana
);

-- Entries table: Ek account mein paisa aaya ya gaya
CREATE TABLE "entries" (
  "id" bigserial PRIMARY KEY,
  "account_id" bigint NOT NULL,        -- kis account ka entry hai
  "amount" bigint NOT NULL,            -- kitna paisa (+ve = aaya, -ve = gaya)
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

-- Transfers table: Ek account se doosre mein paisa gaya
CREATE TABLE "transfers" (
  "id" bigserial PRIMARY KEY,
  "from_account_id" bigint NOT NULL,   -- kis account SE gaya
  "to_account_id" bigint NOT NULL,     -- kis account MEIN gaya
  "amount" bigint NOT NULL,            -- kitna paisa (hamesha +ve)
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

-- Foreign Keys: Ye ensure karte hain ki account exist kare
-- Agar koi aisi account_id daalo jo exist nahi karti → ERROR!
ALTER TABLE "entries" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id");
ALTER TABLE "transfers" ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id");
ALTER TABLE "transfers" ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id");

-- Indexes: Searching fast karte hain
-- Jaise book mein index hota hai, waise hi
CREATE INDEX ON "accounts" ("owner");
CREATE INDEX ON "entries" ("account_id");
CREATE INDEX ON "transfers" ("from_account_id");
CREATE INDEX ON "transfers" ("to_account_id");
CREATE INDEX ON "transfers" ("from_account_id", "to_account_id");

-- Comments: Documentation
COMMENT ON COLUMN "entries"."amount" IS 'can be negative or positive';
COMMENT ON COLUMN "transfers"."amount" IS 'must be positive';
```

### DOWN File — `000001_init_schema.down.sql`

Ye file chalegi jab `migrate down` karoge. Sab undo ho jaayega:

```sql
-- IMPORTANT: Order matters!
-- Pehle entries aur transfers delete karo, PHIR accounts
-- Kyun? Kyunki entries aur transfers mein foreign key hai jo accounts ko point karti hai
-- Agar pehle accounts delete karo → ERROR! (children abhi bhi point kar rahe hain)

DROP TABLE IF EXISTS entries;     -- Pehle ye (kyunki ye accounts pe depend karta hai)
DROP TABLE IF EXISTS transfers;   -- Phir ye (ye bhi accounts pe depend karta hai)
DROP TABLE IF EXISTS accounts;    -- Last mein ye (parent table)
```

**Analogy:** Socho ek building hai. Pehle top floor todna padta hai, phir middle, phir ground floor. Seedha ground floor nahi tod sakte jab upar logbaitha hai!

---

## 🚀 Migration Run Karna

### Step 1: Database bana lo (agar nahi banayi)

```bash
docker exec -it postgres12 createdb --username=root --owner=root simple_bank
```

### Step 2: Migrate UP (tables banao)

```bash
migrate -path db/migration \
  -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" \
  -verbose up
```

**Har part ka matlab:**

| Part | Matlab |
|------|--------|
| `-path db/migration` | Migration files yahan hain |
| `-database "postgresql://..."` | Database ka URL |
| `root:secret` | username:password |
| `localhost:5432` | kahan chal raha hai |
| `simple_bank` | database ka naam |
| `sslmode=disable` | SSL band karo (local development mein zaruri!) |
| `-verbose` | Detail mein batao kya ho raha hai |
| `up` | AAGE jaao (UP migration chalao) |

> ⚠️ **SSL Error:** Agar `sslmode=disable` nahi lagaoge toh error aayega:
> `"SSL is not enabled on the server"`
> Kyunki Docker container mein SSL default off hota hai.

### Step 3: Migrate DOWN (tables hatao — undo!)

```bash
migrate -path db/migration \
  -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" \
  -verbose down
```

---

## 📊 schema_migrations Table

Jab pehli baar migration run hoti hai, ek **extra table** automatically ban jaati hai:

```
schema_migrations
┌─────────┬───────┐
│ version │ dirty │
├─────────┼───────┤
│ 1       │ false │
└─────────┴───────┘
```

| Column | Matlab |
|--------|--------|
| `version` | Last migration jo successfully chali (1 = `000001_init_schema`) |
| `dirty` | Kya last migration fail hui? `false` = sab theek, `true` = problem hai! |

**Agar `dirty = true` ho toh?**
- Matlab last migration beech mein fail ho gayi
- Database inconsistent state mein hai
- Manually fix karo, phir migrate karo
- `migrate force <version>` se dirty flag reset kar sakte ho

---

## 🐳 Naye Docker Commands (Is Lecture Mein)

Video mein kuch naye Docker commands bhi dikhaye gaye:

```bash
# Container STOP karo (band karo, delete nahi)
docker stop postgres12

# Container START karo (wapas chalu)
docker start postgres12

# Container PERMANENTLY delete karo
docker rm postgres12

# Container ke shell mein jaao
docker exec -it postgres12 /bin/sh
# Alpine image mein /bin/bash NAHI hota, /bin/sh use karo!

# Shell ke andar database banao
createdb --username=root --owner=root simple_bank

# Shell ke andar database delete karo
dropdb simple_bank

# Shell se bahar aao
exit
```

---

## 📋 Makefile (Commands Short Karo!)

Video mein ek `Makefile` banaya gaya taaki lambe commands baar baar type na karne pade:

```makefile
postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

.PHONY: postgres createdb dropdb migrateup migratedown
```

**Ab bas type karo:**
```bash
make postgres      # PostgreSQL container start
make createdb      # Database banao
make migrateup     # Tables banao
make migratedown   # Tables hatao (undo)
make dropdb        # Database delete karo
```

**`.PHONY` kya hai?** Ye Make ko batata hai ki "ye actual files nahi hain, ye commands hain". Bina iske agar koi file `postgres` naam ki hoti toh Make confuse ho jaata.

---

## 📁 Is Folder Mein Kya Hai

```
lecture_3_db_migration/
├── EXPLANATION.md                         ← Ye file (detailed explanation)
├── db/
│   └── migration/
│       ├── 000001_init_schema.up.sql      ← Tables banane ka SQL
│       └── 000001_init_schema.down.sql    ← Tables delete karne ka SQL
└── Makefile                               ← Short commands
```

---

## 💡 Key Takeaways

1. **Migration = Database ka Git** — changes track karo, undo karo
2. **Har migration mein 2 files** — UP (aage) aur DOWN (peechhe)
3. **Sequential numbering** — 000001, 000002... order mein chalti hain
4. **UP = order mein**, **DOWN = ULTA order mein**
5. **Foreign key ka dhyan rakhna** — pehle children drop karo, phir parent
6. **`sslmode=disable`** zaroori hai local Docker mein
7. **`schema_migrations` table** automatically banti hai — version track karti hai
8. **Makefile se life easy** — lambe commands short ho jaate hain
9. **`dirty = true`** matlab problem hai — manually fix karo pehle
