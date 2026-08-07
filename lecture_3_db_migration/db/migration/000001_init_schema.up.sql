-- ============================================================
-- UP MIGRATION: Ye tables banata hai
-- Jab "migrate up" chalate ho toh ye file execute hoti hai
-- ============================================================

-- ACCOUNTS TABLE
-- Ye bank account store karta hai
-- bigserial = auto-increment integer (1, 2, 3, 4...)
-- PRIMARY KEY = unique identifier, do accounts ka same ID nahi ho sakta
-- NOT NULL = ye field empty nahi chhod sakte
-- DEFAULT (now()) = agar value nahi di toh current time daal do
CREATE TABLE "accounts" (
  "id" bigserial PRIMARY KEY,
  "owner" varchar NOT NULL,
  "balance" bigint NOT NULL,
  "currency" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

-- ENTRIES TABLE  
-- Ye ek account mein paisa aana/jaana track karta hai
-- Jaise tumhare bank statement mein dikhta hai:
--   +500 (salary aayi)
--   -100 (kuch kharida)
CREATE TABLE "entries" (
  "id" bigserial PRIMARY KEY,
  "account_id" bigint NOT NULL,
  "amount" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

-- TRANSFERS TABLE
-- Ye ek account se doosre mein paisa transfer track karta hai
-- Example: Apoorv ne Bob ko 500 rupaye bheje
CREATE TABLE "transfers" (
  "id" bigserial PRIMARY KEY,
  "from_account_id" bigint NOT NULL,
  "to_account_id" bigint NOT NULL,
  "amount" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

-- FOREIGN KEYS
-- Ye ensure karte hain ki sirf EXISTING accounts ke entries ban sakein
-- Agar account_id = 999 daalo aur account 999 exist nahi karta → ERROR!
ALTER TABLE "entries" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id");
ALTER TABLE "transfers" ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id");
ALTER TABLE "transfers" ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id");

-- INDEXES
-- Ye searching fast karte hain
-- Bina index ke: database SAARI rows check karega (slow)
-- Index ke saath: directly sahi row pe pahunch jaata hai (fast)
-- Jaise book mein index/table of contents hota hai
CREATE INDEX ON "accounts" ("owner");
CREATE INDEX ON "entries" ("account_id");
CREATE INDEX ON "transfers" ("from_account_id");
CREATE INDEX ON "transfers" ("to_account_id");
CREATE INDEX ON "transfers" ("from_account_id", "to_account_id");

-- COMMENTS
-- Ye sirf documentation hai, koi effect nahi padta code pe
COMMENT ON COLUMN "entries"."amount" IS 'can be negative or positive';
COMMENT ON COLUMN "transfers"."amount" IS 'must be positive';
