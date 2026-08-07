-- ============================================================
-- DOWN MIGRATION: Ye tables delete karta hai (UNDO!)
-- Jab "migrate down" chalate ho toh ye file execute hoti hai
-- ============================================================

-- ORDER BAHUT IMPORTANT HAI!
-- Pehle entries aur transfers drop karo, PHIR accounts
-- 
-- Kyun? Kyunki entries aur transfers mein FOREIGN KEY hai
-- jo accounts table ko point karti hai.
-- 
-- Agar pehle accounts delete karo → PostgreSQL error dega:
-- "cannot drop table accounts because other objects depend on it"
--
-- Socho aise: Pehle bachche (entries, transfers) hatao,
-- phir parent (accounts) hatao!

DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS accounts;
