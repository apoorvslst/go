package util

import (
	"math/rand"
	"strings"
	"time"
)

// alphabet letters jo hum random string banane me use karenge
const alphabet = "abcdefghijklmnopqrstuvwxyz"

// init function package load hote hi automatically run ho jaata hai
func init() {
	// rand.Seed() batata hai ki random numbers kahan se shuru karein.
	// time.Now().UnixNano() se hum ensure karte hain ki har baar alag numbers generate hon.
	rand.Seed(time.Now().UnixNano())
}

// RandomInt ek random number deta hai min aur max ke beech (e.g., balance ke liye)
func RandomInt(min, max int64) int64 {
	// rand.Int63n(max-min+1) ek random number deta hai 0 se leke (max-min) tak.
	// Phir hum usme min add kar dete hain.
	return min + rand.Int63n(max-min+1)
}

// RandomString ek n-length ka random string banata hai (e.g., owner name ke liye)
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	// n baar loop chalake random letters alphabet string me se pick karta hai
	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c) // aur ek naye string me append karta jaata hai
	}

	return sb.String()
}

// RandomOwner ek 6-characters ka lamba random naam generate karta hai
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney ek random amount (0 se 1000 ke beech) balance ke liye generate karta hai
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency ek valid currency code (EUR, USD, ya CAD) return karta hai randomly
func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	n := len(currencies)
	return currencies[rand.Intn(n)] // Koi ek currency uthake wapis deta hai
}
