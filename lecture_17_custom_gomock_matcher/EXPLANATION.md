# 👤 Lecture 17: Write Custom Gomock Matcher for Password Testing

## 🎯 High Level Overview

Is lecture mein humne seekha ki kaise `CreateUser` API ka unit test likhte waqt humein ek special challenge milta hai: **Bcrypt hashing ka random salt**.
Kyunki bcrypt har baar ek naya random salt generate karta hai, same password ka hash hamesha alag aata hai. Is wajah se, standard `gomock.Eq()` matcher fail ho jata hai (kyunki database ka expected hash aur handler mein generate hua naya hash string level par alag dikhte hain).

Is problem ko solve karne ke liye humne ek **Custom Matcher** banaya, `gomock.Matcher` interface ko implement karke.

### Naye concepts jo humne seekhe:
1. **Gomock Matcher Limitation**: `gomock.Any()` test ko weak bana deta hai, aur `gomock.Eq()` random salts par fail ho jata hai.
2. **Custom Matcher Implementation**: Apna khud ka matcher `eqCreateUserParamsMatcher` banaya jo specifically hashed passwords ko `util.CheckPassword` use karke match karta hai.

---

## 📝 Custom Matcher Banane Ka Reason

Agar hum `gomock.Any()` use karte hain:
- Test pass ho jayega bhale hi API handler koi input parameters pass hi na kare (e.g., empty values bhej de). Ye test bohot weak hota hai kyunki wo implementation ke flaws catch nahi karta.

Agar hum `gomock.Eq()` use karte hain:
- Handler API `CreateUser` ko database mein save karne se pehle password ko hash karti hai.
- Test setup me hum bhi test case ke liye same password ka ek hash banate hain expect karne ke liye.
- Dono hashes same plain password ke hote hain, par random salt ki wajah se string values alag alag hoti hain. `gomock.Eq()` unhe match nahi kar pata aur Test fail ho jata hai!

---

## 🛠️ Step 1: Custom Matcher Struct Banana

Custom matcher banane ke liye humein `gomock.Matcher` interface implement karna padta hai, jisme do functions hote hain: `Matches` aur `String`.

Sabse pehle humne ek naya struct banaya `api/user_test.go` mein (Mock store expectation test jahan hota hai):

```go
type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams // Expected arguments (Username, Email, etc.)
	password string              // Plaintext password 
}
```
**Explanation:**
- `arg`: Isme expected username, email, full_name hoga.
- `password`: Isme plain naked password pass karenge taaki baad mein bcrypt ki `CheckPassword` function call karke manually check kar sakein ki handler se aaya hua hash is plain password se match karta hai ya nahi.

---

## 🛠️ Step 2: `Matches` Function Implement Karna

Ye function decide karega ki actual input argument (jo handler ne generate kiya) hamare expected argument se match kar raha hai ya nahi.

```go
func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	// 1. Argument ko db.CreateUserParams mein convert karo
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	// 2. Password verify karo bcrypt logic se
	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false // Password galat nikla!
	}

	// 3. Ab HashedPassword ka issue fix ho gaya, toh dono ko artificially same bana do
	e.arg.HashedPassword = arg.HashedPassword

	// 4. Baaki saari fields (Username, Email, etc.) compare karo using DeepEqual
	return reflect.DeepEqual(e.arg, arg)
}
```

**Explanation:**
1. Pehle hum type cast karte hain `x` ko `db.CreateUserParams` mein.
2. Uske baad sabse important step: Hum apne plain password ko actual argument ke `HashedPassword` ke saath `CheckPassword` function me dalte hain. (Bcrypt khud salt wala check sambhal lega).
3. Agar password sahi hai, toh hum `e.arg.HashedPassword` mein bhi wahi actual hash daal dete hain, taaki jab niche `reflect.DeepEqual` chale, toh hashes wali unequal string fail na kare. Baaki fields automatically strictly match ho jayengi.

---

## 🛠️ Step 3: `String` aur Helper Function Banana

Jab test fail ho, toh logs acche se dikhane ke liye `String()` method implement kiya:

```go
func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}
```

Aur finally, ek helper function banaya jo is matcher ka instance return karega, taaki hum isko easily `gomock` setup me call kar sakein:

```go
func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{
		arg:      arg,
		password: password,
	}
}
```

---

## 🛠️ Step 4: Test mein Custom Matcher Use Karna

Purana code jahan humne `gomock.Any()` likha tha, usko ab naye function se replace karenge:

```go
// Pehle Test ka logic
arg := db.CreateUserParams{
    Username: user.Username,
    FullName: user.FullName,
    Email:    user.Email,
}

// ❌ WEAK TEST (Purana Code)
// store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(1).Return(user, nil)

// ✅ STRONG TEST (Naya Code with Custom Matcher)
store.EXPECT().
	CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
	Times(1).
	Return(user, nil)
```

---

## 🚀 Summary (Faayda Kya Hua?)
- **Strong Unit Tests:** Ab hamara API test input request ko strictly validate karta hai. Agar kisi developer ne handler mein galti se user ka password empty bhej diya, input params ignore kar diye, ya constant `"xyz"` password hash bhej diya, toh test immediately fail ho jayega.
- **Bcrypt Workaround:** Custom matcher banane ki wajah se salt ka random nature unit tests mein problem create nahi karta. Hum hashes ko direct string-match karne ke bajaye logical bcrypt verification karte hain.
