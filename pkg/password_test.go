package utils

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"
	hash, err := HashPassword(password)

	assert.NoError(t, err, "HashPassword should not return an error")
	assert.NotEmpty(t, hash, "Hash should not be empty")
	assert.NotEqual(t, password, hash, "Hash should be different from the original password")
}

func TestCheckPasswordHash_Success(t *testing.T) {
	password := "mysecretpassword"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	match := CheckPasswordHash(password, hash)
	assert.True(t, match, "Password and hash should match")
}

func TestCheckPasswordHash_Failure(t *testing.T) {
	password := "mysecretpassword"
	wrongPassword := "wrongpassword"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	match := CheckPasswordHash(wrongPassword, hash)
	assert.False(t, match, "Wrong password should not match the hash")
}

func TestCheckPasswordHash_InvalidHash(t *testing.T) {
	password := "mysecretpassword"
	invalidHash := "this-is-not-a-valid-bcrypt-hash"

	match := CheckPasswordHash(password, invalidHash)
	assert.False(t, match, "Password should not match an invalid hash")
}
