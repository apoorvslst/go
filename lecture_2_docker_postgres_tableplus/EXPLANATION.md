# 🐳 Lecture 2: Docker + PostgreSQL + TablePlus Setup

## Ye Lecture Kya Hai?

Is lecture mein 3 cheezein sikhayi gayi hain:
1. **Docker Desktop** install karna aur use karna
2. **PostgreSQL** container chalana Docker mein
3. **TablePlus** se database ko visually manage karna

---

## 🐋 Part 1: Docker Kya Hai? (Bahut Simple!)

Socho Docker ek **dabba (box)** hai. Agar tum directly apne computer pe PostgreSQL install karo, toh bahut gadbad ho sakti hai — versions clash karenge, settings bigadh jayengi, uninstall karna mushkil hoga.

Docker kya karta hai? Woh PostgreSQL ko ek **alag dabba** (container) mein rakh deta hai. Tumhara computer clean rahega, aur jab chaaho container delete kar do — koi problem nahi!

### Image vs Container — Samjho aise:

```
IMAGE = Recipe (frozen pizza ka dabba)
  └── Ek blueprint hai. Isse kuch nahi hota directly.
  
CONTAINER = Running instance (pizza jo tune bake kar diya)
  └── Ye actual cheez hai jo chal rahi hai.
```

**Example:** Ek image se 5 containers bana sakte ho — jaise ek recipe se 5 pizza bana sakte ho!

---

## 🔧 Part 2: Docker Commands (Step by Step)

### Step 1: Image Download karo (Pull)

```bash
docker pull postgres:12-alpine
```

**Kya ho raha hai?**
- `docker pull` = image download karo
- `postgres` = PostgreSQL ka image
- `:12-alpine` = version 12, Alpine Linux pe (bahut chhota ~150MB)
- Alpine = ek bahut lightweight Linux hai, normal Linux ~1GB hota hai, Alpine ~150MB

### Step 2: Container Chalao (Run)

```bash
docker run --name postgres12 \
  -p 5432:5432 \
  -e POSTGRES_USER=root \
  -e POSTGRES_PASSWORD=secret \
  -d postgres:12-alpine
```

**Har flag ka matlab:**

| Flag | Kya karta hai | Example |
|------|--------------|---------|
| `--name postgres12` | Container ka naam rakh do | Jaise pet ka naam rakhte ho |
| `-p 5432:5432` | Port mapping (bridge) | Tumhara port 5432 → container ka 5432 |
| `-e POSTGRES_USER=root` | Username set karo | Login ke liye |
| `-e POSTGRES_PASSWORD=secret` | Password set karo | Login ke liye |
| `-d` | Background mein chalao | Terminal free rahega |

### Port Mapping Kya Hai? (BAHUT IMPORTANT!)

```
Ye socho:

Tumhara Computer                    Docker Container
┌──────────────────┐                ┌──────────────────┐
│                  │                │                  │
│  Go app          │     BRIDGE     │   PostgreSQL     │
│  TablePlus       │◄──────────────►│   Port 5432      │
│  Browser         │   (p 5432:5432)│                  │
│                  │                │                  │
│  localhost:5432  │                │  container:5432  │
└──────────────────┘                └──────────────────┘
```

Docker container ek **alag network** mein hota hai. Bina port mapping ke tumhara computer container se baat nahi kar paayega. `-p 5432:5432` ek **bridge/pul** bana deta hai.

### Step 3: Check karo sab chal raha hai

```bash
# Saare running containers dekho
docker ps

# Output kuch aisa dikhega:
# CONTAINER ID   IMAGE                  STATUS         PORTS                    NAMES
# abc123def      postgres:12-alpine     Up 5 minutes   0.0.0.0:5432->5432/tcp   postgres12
```

### Step 4: Container ke andar jaao

```bash
# PostgreSQL console mein jaao directly
docker exec -it postgres12 psql -U root

# Kya matlab hai:
# exec     = container ke andar command chalao
# -it      = interactive mode (tum type kar sako)
# psql     = PostgreSQL ka CLI tool
# -U root  = root user se login
```

**Note:** Password nahi maangega kyunki PostgreSQL by default local connections pe **trust** karta hai — matlab localhost se aao toh password ki zaroorat nahi.

### Step 5: Database banao

```bash
# Container ke shell mein jaao
docker exec -it postgres12 /bin/sh

# Shell ke andar, database banao
createdb --username=root --owner=root simple_bank

# Database console mein jaao
psql simple_bank

# Kuch try karo
SELECT now();  -- current time dikhayega

# Bahar aao
\q     -- psql se bahar
exit   -- shell se bahar
```

### Docker Commands Cheat Sheet (Yaad rakhne wale!)

```bash
# ========= IMAGES =========
docker images                    # Saari images dikha do
docker pull postgres:12-alpine   # Image download karo

# ========= CONTAINERS =========
docker ps                        # Running containers
docker ps -a                     # SAARE containers (band bhi)
docker run ...                   # Naya container chalao
docker start postgres12          # Band container chalu karo
docker stop postgres12           # Chalta container band karo
docker rm postgres12             # Container delete karo (pehle stop karo)

# ========= INSIDE CONTAINER =========
docker exec -it postgres12 psql -U root    # PostgreSQL console
docker exec -it postgres12 /bin/sh         # Container ka shell

# ========= DEBUGGING =========
docker logs postgres12           # Container ke logs dekho (kya hua andar)
```

---

## 🖥️ Part 3: TablePlus Setup

TablePlus ek **GUI tool** hai — matlab buttons aur click karke database manage karo, SQL terminal mein type karne ki zaroorat nahi.

### Connection Settings (Video mein ye use kiye):

| Setting | Value | Kyun? |
|---------|-------|-------|
| **Host** | `localhost` | Container tumhare computer pe chal raha hai |
| **Port** | `5432` | Default PostgreSQL port (humne map bhi kiya) |
| **User** | `root` | Humne `-e POSTGRES_USER=root` set kiya tha |
| **Password** | `secret` | Humne `-e POSTGRES_PASSWORD=secret` set kiya tha |
| **Database** | `root` | Jab tak explicitly nahi bataoge, username = database name |

### TablePlus mein kya kya kar sakte ho:

1. **Tables dekho** — Left sidebar mein table names dikhte hain, click karo
2. **Data dekho** — Table pe click karke saara data dikh jaata hai
3. **Structure dekho** — "Structure" tab mein columns, types, constraints dikhte hain
4. **SQL queries chalao** — SQL icon pe click karo, query likho, Ctrl+Enter
5. **Data edit karo** — Cell pe double click karke value change karo
6. **Changes save karo** — Ctrl+S (jab tak save nahi karoge, changes apply nahi honge)

### Simple Bank ka Schema chalana

Video mein SQL file open karke saari queries select karke run ki gayi thi:

```sql
-- Ctrl+A se sab select karo, phir Ctrl+Enter se run karo
-- Ye 3 tables ban jayengi: accounts, entries, transfers
```

Uske baad `Ctrl+R` (Refresh) karo → tables left side mein dikh jayengi!

### NOT NULL Fix

Video mein pehle foreign key columns **nullable** the (matlab empty ho sakte the). But bank mein har entry ko ek account se linked hona chahiye — toh `NOT NULL` add kiya gaya:

```sql
-- Pehle (GALAT - nullable tha):
"account_id" bigint,          -- NULL ho sakta hai ❌

-- Baad mein (SAHI - not null):  
"account_id" bigint NOT NULL,  -- NULL nahi ho sakta ✅
```

---

## 📝 Is Folder Mein Kya Hai

```
lecture_2_docker_postgres_tableplus/
├── EXPLANATION.md          ← Ye file (jo tum padh rahe ho)
└── docker_commands.sh      ← Saare Docker commands ek jagah
```

---

## 💡 Key Takeaways (Yaad Rakhne Wali Cheezein)

1. **Docker = dabba** — har cheez alag alag dabbon mein rakhte hain
2. **Image = recipe**, **Container = running app**
3. **Port mapping (`-p`)** = bridge banata hai tumhare computer aur container ke beech
4. **`docker exec -it`** = container ke andar ghusne ka raasta
5. **TablePlus** = database ko GUI se manage karne ka tool (life easy!)
6. **`docker ps`** = sabse useful command — kya chal raha hai dekho
7. **`docker logs`** = kuch gadbad ho toh logs dekho
