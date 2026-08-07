-- ============================================================
-- SQLC Query File: account.sql
-- 
-- Ye file SQLC ka INPUT hai
-- SQLC isse padhke Go code generate karta hai
--
-- SYNTAX:
--   -- name: FunctionName :return_type
--   SQL QUERY;
--
-- Return types:
--   :one  → 1 row return hogi  → func returns (Account, error)
--   :many → bahut rows         → func returns ([]Account, error)
--   :exec → kuch nahi          → func returns error
-- ============================================================

-- name: CreateAccount :one
-- Naya account banao. $1, $2, $3 = parameters (Go se aayenge)
-- RETURNING * = INSERT ke baad bana hua row wapas do
INSERT INTO accounts (owner, balance, currency)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAccount :one
-- Ek account ID se dhoondho
-- LIMIT 1 = sirf pehla result do
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: ListAccounts :many
-- Saare accounts do, but pagination ke saath
-- LIMIT = kitne results chahiye
-- OFFSET = kitne skip karo (page 2 ke liye offset=10)
SELECT * FROM accounts
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateAccount :one
-- Account ka balance update karo
UPDATE accounts
SET balance = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
-- Account delete karo. :exec kyunki kuch return nahi chahiye
DELETE FROM accounts
WHERE id = $1;

-- ============================================================
-- ENTRY QUERIES (agar entries table ke liye bhi chahiye)
-- ============================================================

-- name: CreateEntry :one
INSERT INTO entries (account_id, amount)
VALUES ($1, $2)
RETURNING *;

-- name: GetEntry :one
SELECT * FROM entries
WHERE id = $1 LIMIT 1;

-- name: ListEntries :many
SELECT * FROM entries
WHERE account_id = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- ============================================================
-- TRANSFER QUERIES
-- ============================================================

-- name: CreateTransfer :one
INSERT INTO transfers (from_account_id, to_account_id, amount)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers
WHERE id = $1 LIMIT 1;

-- name: ListTransfers :many
SELECT * FROM transfers
WHERE from_account_id = $1 OR to_account_id = $1
ORDER BY id
LIMIT $2
OFFSET $3;
