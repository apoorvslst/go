# 🔄 Lecture 10: Continuous Integration with GitHub Actions

## 🌟 High Level Overview (Ye Lecture Kya Hai?)
Continuous Integration (CI) ek software development process hai jahan code changes automatically build aur test hote hain jab bhi koi team member code push karta hai.
Is lecture mein humne setup kiya hai **GitHub Actions** taaki hamara Golang aur Postgres wala Simple Bank project automatically test ho sake jab bhi GitHub pe naya code aaye.

### ⚙️ GitHub Actions Basics
- **Workflow**: Ek automated procedure (jaise build-and-test) jo kisi event par chalta hai.
- **Trigger**: Woh event jo workflow ko start kare (jaise `push`, `pull_request`, ya schedule).
- **Job**: Ek set of steps jo kisi ek server (runner) par chalte hain. By default saare jobs parallel mein chalte hain.
- **Runner**: Ek machine (server) jo aapka job execute karta hai (e.g., Ubuntu).
- **Step**: Individual task jo ek ke baad ek (serially) chalta hai ek job ke andar.
- **Action**: Ek pre-written command ya script jo aap baar-baar reuse kar sakte hain.

---

## 🛠️ Step-by-Step Implementation

### Step 1: Workflow File Banana
GitHub ko batane ke liye ki CI chalana hai, humein ek YAML file banani padti hai ek specific folder structure mein.
1. Apne project root mein folder banao: `.github/workflows/`
2. Uske andar ek file banao: `ci.yml`

### Step 2: Trigger aur Runner Define Karna
Open `ci.yml` aur basic setup likhte hain:

```yaml
name: ci-test

on:
  push:
    branches: [ master ]
  pull_request:
    branches: [ master ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
```

**📝 Code Explanation:**
- `name: ci-test` -> GitHub Actions ke dashboard pe ye naam dikhega.
- `on: push/pull_request` -> Ye workflow tab trigger hoga jab bhi `master` branch par code push hoga ya PR aayega.
- `jobs: test` -> Hum ek job bana rahe hain jiska naam `test` hai.
- `runs-on: ubuntu-latest` -> Hum GitHub se ek naya Ubuntu server maang rahe hain jisme hamara code chalega.

---

### Step 3: Postgres Database Service Setup Karna
Hamare unit tests ko chalne ke liye ek real Postgres database chahiye hota hai. Toh hum Ubuntu runner ke andar ek Postgres ki background service chalayenge.

```yaml
    services:
      postgres:
        image: postgres:12
        env:
          POSTGRES_USER: root
          POSTGRES_PASSWORD: secret
          POSTGRES_DB: simple_bank
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
```

**📝 Code Explanation:**
- `services: postgres` -> Runner ke andar background mein Postgres chalao.
- `image: postgres:12` -> Postgres ka version 12 Docker image use karo.
- `env:` -> Username, password, aur db ka naam bilkul waise hi set kiya jaise local setup mein tha.
- `ports: - 5432:5432` -> **Important!** Container ke port 5432 ko Ubuntu runner ke port 5432 se map karo warna Go code DB se connect nahi kar payega.
- `options: --health-cmd pg_isready` -> Runner wait karega jab tak database properly start aur ready na ho jaye agle steps chalane se pehle.

---

### Step 4: Steps (Actions) Define Karna
Ab hum Ubuntu runner ko batayenge ki kya kya actions lene hain.

```yaml
    steps:
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.15

    - name: Check out code
      uses: actions/checkout@v2
```

**📝 Code Explanation:**
- `uses: actions/setup-go@v2` -> GitHub ka pre-written action use karke Ubuntu server pe Go install karo (version 1.15).
- `uses: actions/checkout@v2` -> Hamari Github repository se code ko is Ubuntu server pe download/checkout karo.

---

### Step 5: golang-migrate CLI Install Karna
Hamare tests bina database tables (schema) ke nahi chalenge. Tables banane ke liye humein `make migrateup` chalana hai, jiske liye `migrate` tool chahiye, jo runner mein pehle se nahi hota.

```yaml
    - name: Install golang-migrate
      run: |
        curl -L https://github.com/golang-migrate/migrate/releases/download/v4.12.2/migrate.linux-amd64.tar.gz | tar xvz
        sudo mv migrate /usr/bin/
        which migrate
```

**📝 Code Explanation:**
- `run: |` -> Pipe symbol `|` batata hai ki hum multiple lines ki bash command likhne wale hain.
- `curl -L ... | tar xvz` -> golang-migrate ki linux binary internet se download karo aur extract karo.
- `sudo mv migrate /usr/bin/` -> Download ki hui `migrate` file ko `/usr/bin/` folder mein bhejo aur uska naam `migrate` rakho. Isse ye system-wide available ho jayegi. `sudo` isliye use kiya kyunki system folder modify kar rahe hain.
- `which migrate` -> Test karne ke liye ki tool successfully install hua ya nahi.

---

### Step 6: Migrations aur Tests Run Karna
Ab sab kuch setup hai (Go hai, Code hai, DB chal raha hai, Migrate tool hai). Final step!

```yaml
    - name: Run migrations
      run: make migrateup

    - name: Test
      run: make test
```

**📝 Code Explanation:**
- `make migrateup` -> Database mein saari tables banayega.
- `make test` -> Hamare Golang ke saare unit tests run karega. Agar tests pass hue, toh GitHub pe Green Tick (✅) aayega, warna Red Cross (❌).

---

## 🚀 Final Summary (Troubleshooting)
Video mein isko setup karte waqt kuch errors aaye jinse humne seekha:

1. **Connection Refused Error**: Kyunki humne Postgres DB configure nahi kiya tha pehle. Solved by adding `services`.
2. **migrate command not found**: Kyunki `golang-migrate` tool runner pe missing tha. Solved by `curl` script.
3. **migrate binary name issue**: File extract hone ke baad naam `migrate.linux-amd64` tha. `sudo mv` karte waqt humne explicitly naam `migrate` rakha taaki Makefile usko dhoondh sake.
4. **Port not exposed**: DB running tha but connection fail hua. Solved by adding `ports: - 5432:5432` to the postgres service.

Done! Ab aapka CI pipeline ready hai! 🚀
