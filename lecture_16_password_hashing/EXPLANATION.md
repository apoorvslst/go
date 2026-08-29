# 👤 Lecture 16: Securely Store Passwords - Hash Password in Go with Bcrypt

## 🎯 High Level Overview

Is lecture mein humne sikha ki passwords ko plain text mein database mein store karna kyun dangerous hai aur kaise hum `bcrypt` hashing algorithm ka use karke passwords ko securely store kar sakte hain.

### Naye features jo humne add kiye:
1. **Password Hashing**: Naya package `golang.org/x/crypto/bcrypt` install kiya.
2. **Util Functions**: `util` folder ke andar `password.go` file banayi jisme passwords hash karne aur verify karne ke functions likhe.
3. **Unit Tests**: `password_test.go` banaya taaki ensure kar sakein ki humara hashing logic sahi se kaam kar raha hai (including matching passwords, non-matching passwords, aur random salt verify karna).

---

## 📝 Correct Order of Writing Code (Agar Khud Se Karna Ho)

Agar aap isko khud implement kar rahe hain, toh ye steps follow karein:

### Phase 1: Package Installation & Utility Functions
1. **Terminal** — Go crypto package download karein (`go get golang.org/x/crypto/bcrypt`).
2. **util/password.go** — `HashPassword(password string)` aur `CheckPassword(password string, hashedPassword string)` functions likhein.

### Phase 2: Testing the Implementation
3. **util/password_test.go** — Unit tests likhein jo normal scenario, wrong password scenario, aur hash ki uniqueness check karein.
4. **Terminal** — `go test ./util/...` chala kar verify karein.

---

## 🛠️ Step 1: Install Bcrypt Package

Go mein passwords hash karne ke liye standard library ka official `crypto` package use hota hai.

Terminal mein command run karein:
```bash
go get golang.org/x/crypto/bcrypt
```
**Explanation:**
- `bcrypt` ek one-way hashing algorithm hai jisme inbuilt "salt" use hota hai taaki same passwords ka hash bhi alag-alag aaye. Ise crack karna (Brute-force/Rainbow tables) bahut mushkil hai.

---

## 🛠️ Step 2: Hashing & Verification Logic (`util/password.go`)

Ek nayi file banayein `util/password.go` jisme ye functions honge:

```go
package util

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a plaintext password and returns its bcrypt hash
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost is usually 10. Higher cost means more security but slower execution.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword checks if the provided password is correct or not
func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
```

**Explanation:**
- `GenerateFromPassword`: Password string ko bytes (`[]byte`) mein convert karta hai aur ek random salt add karke hash generate karta hai. `DefaultCost` (10) decide karta hai ki hashing kitni CPU-intensive hogi.
- `CompareHashAndPassword`: Ye purane generated hash ko aur user ke naye plaintext password ko compare karta hai. Ye directly match nahi karta, balki purane hash se "salt" nikal kar naye password ko bhi wahi salt dekar hash karta hai aur phir result check karta hai.

---

## 🛠️ Step 3: Unit Testing (`util/password_test.go`)

Ab hume verify karna hai ki logic theek se kaam kar raha hai ya nahi. Isliye `util/password_test.go` file banayenge:

```go
package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	password := RandomString(6) // Jo RandomString util pehle banaya tha

	// Test 1: Generate Hash correctly
	hashedPassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)

	// Test 2: Check matching password
	err = CheckPassword(password, hashedPassword1)
	require.NoError(t, err)

	// Test 3: Check wrong password
	wrongPassword := RandomString(6)
	err = CheckPassword(wrongPassword, hashedPassword1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	// Test 4: Ensure every hash is unique even for same password
	hashedPassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword2)
	require.NotEqual(t, hashedPassword1, hashedPassword2)
}
```

**Explanation:**
- **Test 3 (Wrong Password):** Ensure karta hai ki galat password par `ErrMismatchedHashAndPassword` aana chahiye.
- **Test 4 (Unique Hashes):** Sabse important! Bcrypt har baar ek naya random salt generate karta hai. Iska matlab same password ke liye bhi hashes alag alag honge. Isse hackers database breach hone par same hash dekh kar commonly used passwords guess nahi kar sakte.

---

## 🚀 Summary (Faayda Kya Hua?)
- **Security First:** Humara simple bank application ab secure hai kyunki database admin ya hackers kisi ko bhi real password nahi dikhega.
- **Salt Generation:** Hamein alag se salt banakar store nahi karna pada, bcrypt hash string ke andar hi usko encode karke rakhta hai.
- **Robust Tests:** Humne edge cases (wrong password, unique hashes) ke liye unit tests bhi likh liye hain, jo CI pipelines mein future code breaks rokega.
