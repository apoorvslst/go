# 🔄 Lecture 12: Load Configuration using Viper

## 🎯 High Level Overview
Jab hum real-world applications banate hain, toh unhe alag-alag environments me chalana padta hai jaise **Development, Testing, Staging, aur Production**. Har environment ki setting (jaise Database URL, Server Port) alag hoti hai. Hum in settings ko code ke andar "hardcode" nahi kar sakte kyunki code change kiye bina settings badalna zaroori hota hai (jaise Docker container me).

Is problem ko solve karne ke liye hum **Viper** package ka use karte hain. Viper hume madad karta hai configurations ko:
1. File (`app.env`, `config.json`, etc.) se read karne me (Local development ke liye best hai).
2. **Environment Variables** se read karne me (Production/Docker ke liye best hai, jo file ko override kar deta hai).

---

## 🛤️ Flow (Step-by-Step Kya Karna Hai)

1. **Viper Install Karein**: `go get github.com/spf13/viper` chala kar project me Viper package add karein.
2. **Config File Banayein (`app.env`)**: Project root me `.env` format ki file banayein jisme humari default configuration hogi.
3. **Config Loader Likhein (`util/config.go`)**: Ek `Config` struct aur uske andar `LoadConfig` function banayein jo Viper ka use karke config load karega.
4. **`main.go` Update Karein**: Hardcoded constants (dbDriver, dbSource, serverAddress) ko delete karke `LoadConfig(".")` use karein.
5. **`main_test.go` Update Karein**: Testing ke time config parent directory me hoti hai, isliye waha `LoadConfig("../..")` use karein.

---

## 🛠️ File 1: `app.env` (Configuration Variables)

Yeh file humare project ke root folder me rakhi jayegi aur isme hum default settings define karenge.

```env
DB_DRIVER=postgres
DB_SOURCE=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable
SERVER_ADDRESS=0.0.0.0:8080
```
> [!NOTE]
> Yahan variable names UPPERCASE me hone chahiye aur equal to (`=`) ke baad bina quotes ke value dalte hain.

> [!TIP]
> **🟢 Node.js Equivalent: .env file**
> Node.js me bhi hum bilkul aise hi `.env` file banate hain aur `dotenv` package (jaise `require('dotenv').config()`) ka use karte hain configuration load karne ke liye.

---

## 🛠️ File 2: `util/config.go` (Config Loader logic)

Is file me hum Viper ka setup karenge config load karne ke liye.

```go
package util

import "github.com/spf13/viper"

// Config struct me saare configuration variables honge
type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}
```
**Explanation:** 
`Config` struct ek container hai jisme hum apni variables ko save karenge. 
`mapstructure` tags Viper (jo mapstructure package internally use karta hai) ko batate hain ki config file ka kaunsa variable Go struct ke kis field me aayega.

```go
// LoadConfig file ya env var se configurations read karta hai
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path) // Viper ko config file ka rasta (folder) batao
	viper.SetConfigName("app") // Config file ka naam (bina extension ke)
	viper.SetConfigType("env") // Config file ka format type
```
**Explanation:** 
- `AddConfigPath(path)`: Jis folder se LoadConfig call hoga, waha config dhundega.
- `SetConfigName("app")`: File ka naam `app` hoga.
- `SetConfigType("env")`: Extension `.env` hoga (json, toml kuch bhi ho sakta hai).

```go
	viper.AutomaticEnv() // Agar environment variables defined hain, toh unko file variables par override kardo
```
**Explanation:** 
Yeh sabse important function hai. Agar hum Docker se chalayenge aur humne environment variable set kiya hoga, toh Viper `app.env` wali value ko ignore karke naye variable ki value le lega.

```go
	err = viper.ReadInConfig() // File read karna start karo
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config) // Data nikal kar Config struct me bhar do
	return
}
```
**Explanation:**
`ReadInConfig()` asal me file ko system se padhta hai. `Unmarshal` uska data le kar humare Go `Config` struct me fit kar deta hai (mapstructure tags ke hisab se).

> [!TIP]
> **🟢 Node.js Equivalent: Config Module**
> ```javascript
> require('dotenv').config();
> 
> const config = {
>   DBDriver: process.env.DB_DRIVER,
>   DBSource: process.env.DB_SOURCE,
>   ServerAddress: process.env.SERVER_ADDRESS
> };
> module.exports = config;
> ```

---

## 🛠️ File 3: `main.go` (Using the new Config)

```go
// Pehle wale const variables yaha se hata diye gaye hain
func main() {
	// Root folder (".") se config load karo
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// Hardcoded DB_DRIVER ki jagah config.DBDriver use karo
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	// Server address ab config.ServerAddress se aayega
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
```
**Explanation:**
Ab humara server hardcoded URL par chalne ke bajaye, seedha `app.env` se port (jaise `0.0.0.0:8080`) aur database URL padhega. Agar main kal terminal me `export SERVER_ADDRESS=0.0.0.0:8081` run karta hu, toh yeh naye port par start hoga bina code change kiye!

> [!TIP]
> **🟢 Node.js Equivalent: Server Start**
> ```javascript
> const config = require('./config');
> app.listen(config.ServerAddress, () => {
>   console.log("Server listening...");
> });
> ```

---

## 🛠️ File 4: `db/sqlc/main_test.go` (Testing with Config)

```go
func TestMain(m *testing.M) {
	// Test file "db/sqlc" folder me hai, par config file root me hai. 
	// Toh hum "parent ke parent" ("../..") directory denge
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
```
**Explanation:**
Tests run karte samay current directory wohi hoti hai jahan test file hai (`db/sqlc/`). Isliye Viper ko batana padta hai ki 2 steps piche (`../..`) ja kar `app.env` dhundo!

---
**Conclusion:**
Viper use karne se humari app **environment-agnostic** ban gayi hai. Hum apne same code ko local machine par alag settings se test kar sakte hain aur Production me alag! 🚀
