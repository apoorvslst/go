package util

import(
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T){
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
