# 👤 Lecture 18: Token-Based Authentication - Implement JWT and PASETO in Go

## 🎯 High Level Overview

Is lecture mein humne sikha ki access tokens (JWT aur PASETO) ko Go mein kaise implement karte hain. Humne dekha ki PASETO implement karna JWT ke comparison mein kitna aasan aur simple hai, aur isse code mein switch karna kitna easy ho jata hai agar design theek se kiya gaya ho.

### Naye features jo humne add kiye:
1. **Token Maker Interface**: Ek general interface banaya taaki hum JWT aur PASETO donon implementations ke beech easily switch kar sakein.
2. **Token Payload**: Ek `Payload` struct design kiya jo token ke andar ka main data (Username, IssuedAt, ExpiredAt, ID) store karta hai. UUID package ka use karke har token ko ek unique ID di gayi.
3. **JWT Maker Implementation**: Go mein `jwt-go` package ka use karke JSON Web Token creation aur verification logic likha. "None" algorithm attack se bachne ke liye verification steps ko carefully handle kiya.
4. **PASETO Maker Implementation**: `o1egl/paseto` package (Version 2) ka use karke PASETO Maker banaya, jo symmetric encryption (Chacha Poly) use karta hai aur JWT se kaafi simple hai.
5. **Unit Tests**: JWT aur PASETO donon makers ke liye extensive unit tests likhe (Happy case, expired token case, aur invalid token algorithm case).

---

## 📝 Correct Order of Writing Code (Agar Khud Se Karna Ho)

Agar aap isko khud implement kar rahe hain, toh ye steps follow karein:

### Phase 1: Interfaces and Payload Logic
1. **token/maker.go** — `Maker` interface declare karein jisme `CreateToken` aur `VerifyToken` methods honge.
2. **Terminal** — UUID package install karein: `go get github.com/google/uuid`.
3. **token/payload.go** — `Payload` struct aur `NewPayload` function likhein. Is file mein `ErrExpiredToken` aur `ErrInvalidToken` errors bhi define karein aur `Valid()` function implement karein jo jwt claims interface ke liye zaroori hai.

### Phase 2: JWT Implementation & Testing
4. **Terminal** — JWT package install karein: `go get github.com/dgrijalva/jwt-go` (ya latest equivalent package).
5. **token/jwt_maker.go** — `JWTMaker` struct banayein aur interface methods ko symmetric key (`HS256`) ke saath implement karein. Verify method mein explicitly error handling (algorithm match aur expired status check) implement karein.
6. **token/jwt_maker_test.go** — JWT functionality test karne ke liye unit tests likhein.

### Phase 3: PASETO Implementation & Testing
7. **Terminal** — PASETO package install karein: `go get github.com/o1egl/paseto`.
8. **token/paseto_maker.go** — `PasetoMaker` struct banayein. Isme sirf `paseto.Encrypt` aur `paseto.Decrypt` ka simple code likhna hota hai.
9. **token/paseto_maker_test.go** — PASETO maker logic ke unit tests likhein, jo mostly JWT tests ke identical honge.

---

## 🛠️ Step 1: Token Maker Interface & Payload (`token/maker.go`, `token/payload.go`)

Sabse pehle hum ek interface aur payload struct define karte hain taaki humara token creation generic rahe.

```go
package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Maker is an interface for managing tokens
type Maker interface {
	CreateToken(username string, duration time.Duration) (string, error)
	VerifyToken(token string) (*Payload, error)
}

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload creates a new token payload with a specific username and duration
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}

// Valid checks if the token payload is valid or not
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
```

**Explanation:**
- Interface ka use: Isse future mein switch karna aasan hoga (JWT to PASETO ya vice versa).
- `Valid()` method: JWT package is function ko use karke internally token ki expiry check karta hai.

---

## 🛠️ Step 2: JWT Implementation (`token/jwt_maker.go`)

```go
package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey}, nil
}

func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))
}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
```

**Explanation:**
- `keyFunc`: Yeh JWT token verify karte waqt check karta hai ki signing method HMAC (`HS256`) hi hai ya nahi. Yeh check bohot zaroori hai to prevent "None" algorithm attack.
- Explicit Error Checking: `jwt.ValidationError` ka use karke hum check karte hain ki error exactly `ErrExpiredToken` hai ya nahi.

---

## 🛠️ Step 3: PASETO Implementation (`token/paseto_maker.go`)

PASETO JWT se kaafi simple aur secure hota hai, aur isme error checking utni complex nahi hoti.

```go
package token

import (
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}
	return maker, nil
}

func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
}

func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
```

**Explanation:**
- `paseto.V2` Chacha Poly 1305 (symmetric encryption) use karta hai. Isme key ka size exactly 32 bytes hona chahiye.
- Verification logic mein sirf `Decrypt` call hota hai aur uske baad payload ka `Valid()` check hota hai. JWT jaisa manual header validation aur complex error tracing nahi chahiye.

---

## 🚀 Summary (Faayda Kya Hua?)
- **Loose Coupling:** Humne `Maker` interface create kiya, taaki backend code (handlers) ko pata bhi nahi chale ki under-the-hood JWT chal raha hai ya PASETO. Future me easily swap kar sakte hain.
- **Improved Security:** JWT mein hone waale common mistakes ("None" algorithm attack) ko explicitly handle kiya aur tokens ke leak hone par identify karne ke liye unhe UUID (ID) di.
- **Simplicity:** PASETO implement karke dekha ki token authentication mein security achieve karna without complex verification kitna aasan ho sakta hai. PASETO developer-friendly aur secure-by-default hai.
