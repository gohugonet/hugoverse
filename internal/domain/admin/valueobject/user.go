package valueobject

import (
	"bytes"
	crand "crypto/rand"
	"golang.org/x/crypto/bcrypt"
	mrand "math/rand"
	"time"
)

type User struct {
	Id    uint64 `json:"id"`
	Email string `json:"email"`
	Hash  string `json:"hash"`
	Salt  string `json:"salt"`
}

func (u *User) Name() string {
	return u.Email
}
func (u *User) ID() uint64 {
	return u.Id
}

var (
	r = mrand.New(mrand.NewSource(time.Now().Unix()))
)

// RandSalt generates 16 * 8 bits of data for a random salt
func RandSalt() ([]byte, error) {
	buf := make([]byte, 16)
	count := len(buf)
	n, err := crand.Read(buf)
	if err != nil {
		return nil, err
	}

	if n != count || err != nil {
		for count > 0 {
			count--
			buf[count] = byte(r.Int31n(256))
		}
	}

	return buf, nil
}

// HashPassword encrypts the salted password using bcrypt
func HashPassword(password, salt []byte) ([]byte, error) {
	salted, err := saltPassword(password, salt)
	if err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword(salted, 10)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// saltPassword combines the salt and password provided
func saltPassword(password, salt []byte) ([]byte, error) {
	salted := &bytes.Buffer{}
	_, err := salted.Write(append(salt, password...))
	if err != nil {
		return nil, err
	}

	return salted.Bytes(), nil
}

// CheckPassword compares the hash with the salted password. A nil return means
// the password is correct, but an error could mean either the password is not
// correct, or the salt process failed - indicated in logs
func CheckPassword(hash, password, salt []byte) error {
	salted, err := saltPassword(password, salt)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword(hash, salted)
}
