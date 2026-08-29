# 👤 Lecture 15: Add Users Table, Foreign Key & Unique Constraints

## 🎯 High Level Overview

Is lecture mein humne apni banking application ke security aur user management ki taraf pehla step liya hai. Abhi tak `accounts` table mein jo `owner` field tha, wo kisi real user se linked nahi tha. Isliye humne ek nayi `users` table banayi aur use `accounts` table se connect kiya.

### Naye features jo humne add kiye:
1. **Users Table**: User ki details (username, password, email, etc.) store karne ke liye. Passwords hamesha hashed form mein store honge security ke liye.
2. **Foreign Key Constraint**: `accounts` table ka `owner` field ab `users` table ke `username` field ka reference dega (1-to-many relationship: 1 user ke multiple accounts ho sakte hain).
3. **Unique Composite Constraint**: Ek user ke paas multiple accounts ho sakte hain, par ek currency ka sirf ek hi account hona chahiye (e.g. 1 USD, 1 EUR allowed hai, par 2 USD accounts nahi). Ise enforce karne ke liye `owner` aur `currency` par unique index banaya.
4. **Database Migrations**: In nayi changes ko properly apply karne ke liye naye migration files banaye aur `Makefile` commands update kiye.

---

## 📝 Correct Order of Writing Code (Agar Khud Se Karna Ho)

Agar aap isko khud implement kar rahe hain, toh ye steps follow karein:

### Phase 1: Database Schema & Migration Generation
1. **dbdiagram.io (Optional)** — `users` table, foreign key, aur unique composite index ko design karo aur uski SQL query generate karo.
2. **Terminal** — Naya migration file generate karo using `migrate create` command.
3. **db/migration/...add_users.up.sql** — Nayi `users` table aur constraints create karne ki SQL queries likho.
4. **db/migration/...add_users.down.sql** — `up.sql` ka exact reverse likho (table and constraints ko drop karne ke liye).

### Phase 2: Resolving Data Conflicts & Makefile Updates
5. **Database Cleanup** — Purana data constraint violate karega, isliye db state ko clean aur reset karo (dirty state = false mark karke `migratedown` chalao).
6. **Makefile** — `migrateup1` aur `migratedown1` commands add karo taaki specific single migrations run ya rollback ki ja sakein.

---

## 🛠️ Step 1: Migration Files Generate Karna

Real-world projects mein hum purani migration file ko delete nahi karte. Database changes track karne ke liye hamesha nayi migration file banti hai.

Terminal mein command run karein:
```bash
migrate create -ext sql -dir db/migration -seq add_users
```
**Explanation:**
- Ye command `db/migration` folder mein do nayi files banayega: `..._add_users.up.sql` aur `..._add_users.down.sql`.
- `-seq`: Sequential prefix (jaise `000002_add_users`) lagane ke liye.

---

## 🛠️ Step 2: `up.sql` (Migration Up Script)

Is file mein hum naya structure banayenge. 

```sql
-- 1. Create the new users table
CREATE TABLE "users" (
  "username" varchar PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

-- 2. Add Foreign Key to existing accounts table
ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");

-- 3. Add Unique Constraint for (owner, currency)
ALTER TABLE "accounts" ADD CONSTRAINT "owner_currency_key" UNIQUE ("owner", "currency");
```
**Explanation:**
- `hashed_password`: Plain text password kabhi save nahi karte.
- `password_changed_at`: Default date '0001-01-01 00:00:00Z' rakhi hai jo ki Go language ki zero time value hoti hai.
- `FOREIGN KEY`: Isse `accounts` table ab without valid `username` koi account banane nahi dega.
- `UNIQUE ("owner", "currency")`: Ye database level pe ensure karta hai ki same owner 2 same currency (e.g. 2 USD) wale accounts create na kar paye.

---

## 🛠️ Step 3: Resolving Dirty State & Cleaning Existing Data

Agar aap directly `make migrateup` chalayenge toh error aayega: `foreign key constraint is violated`.
Kyunki `accounts` table mein already purana kachra (test) data pada hai jinka `owner` nayi `users` table mein exist nahi karta. Failed migration ki wajah se DB "dirty" state mein chala jata hai.

**Fix Karne Ka Tarika (TablePlus ke bina - Using CLI):**
Agar aap TablePlus use nahi kar rahe, toh `schema_migrations` table ko command line (docker exec) ke through update kar sakte hain:
1. Terminal mein ye command run karo taaki dirty state false set ho jaye:
   ```bash
   docker exec -it postgres12 psql -U root -d simple_bank -c "UPDATE schema_migrations SET dirty = false;"
   ```
2. Ab terminal mein `make migratedown` chalao (Isse sab kuch clean ho jayega aur tables drop ho jayengi).
3. Phir `make migrateup` chalao taaki tables shuru se, with users table and constraints, ban jayein.

---

## 🛠️ Step 4: `down.sql` (Migration Down Script)

Jab hume changes rollback karne hote hain, toh `up.sql` mein kiye gaye saare kaam ulte (reverse) sequence mein karne padte hain.

```sql
-- 1. Pehle constraints drop karo
ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "owner_currency_key";
ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";

-- 2. Phir users table ko drop karo
DROP TABLE IF EXISTS "users";
```
**Explanation:**
- Constraint ka naam nikalne ke liye Database IDE (TablePlus) me `accounts` table ke "Info/Structure" section mein ja kar foreign key ka exact name dekh sakte hain (e.g., `accounts_owner_fkey`).
- Pehle constraints hatani zaroori hain, uske baad hi parent table (`users`) ko delete kiya ja sakta hai.

---

## 🛠️ Step 5: `Makefile` Commands Update Karna

Abhi tak hum sirf pure DB ko migrate up ya migrate down kar rahe the. Par specifically sirf last migration ko reverse/apply karne ke liye naye commands banate hain.

**`Makefile` mein add karein:**
```makefile
migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

.PHONY: migrateup1 migratedown1
```
**Explanation:**
- `up 1` aur `down 1`: Ye `migrate` tool ko batata hai ki sirf ek (`1`) step forward ya backward jana hai, na ki saari migration files ko ek sath chala dena hai. Isse debugging aasan hoti hai.

---

## 🚀 Summary (Faayda Kya Hua?)
- **Data Integrity**: Foreign keys se system strict ban gaya hai. Bina exist karne wale user ke ab account create nahi ho sakta.
- **Data Consistency**: Unique composite index ne fix kar diya ki ek user ka ek currency ke liye multiple accounts nahi honge.
- **Migration Rollbacks**: `migrateup1` aur `migratedown1` jaise commands add hone se ab hum schema changes ko smoothly test aur revert kar sakte hain bina poore database ko wipe kiye.
