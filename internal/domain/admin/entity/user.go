package entity

import (
	"bytes"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/admin"
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
	"golang.org/x/crypto/bcrypt"
	mrand "math/rand"
	"time"
)

func (a *Admin) ValidateUser(email, password string) error {
	user, err := getUser(email, a.Repo)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	if err = isUser(user, password); err != nil {
		return err
	}

	return nil
}

func (a *Admin) NewUser(email, password string) (admin.User, error) {
	userData, err := getUser(email, a.Repo)
	if err != nil {
		return nil, err
	}
	if userData != nil {
		return nil, errors.New("user already exists")
	}

	salt, err := randSalt()
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword([]byte(password), salt)
	if err != nil {
		return nil, err
	}

	id, err := a.Repo.NextUserId(email)
	if err != nil {
		return nil, err
	}

	user := &User{
		Id:    id,
		Email: email,
		Hash:  string(hash),
		Salt:  base64.StdEncoding.EncodeToString(salt),
	}

	data, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	err = a.Repo.PutUser(email, data)

	return user, nil
}

type User struct {
	Id    uint64 `json:"id"`
	Email string `json:"email"`
	Hash  string `json:"hash"`
	Salt  string `json:"salt"`
}

func (u *User) Name() string {
	return u.Email
}

var (
	r = mrand.New(mrand.NewSource(time.Now().Unix()))
)

// randSalt generates 16 * 8 bits of data for a random salt
func randSalt() ([]byte, error) {
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

// hashPassword encrypts the salted password using bcrypt
func hashPassword(password, salt []byte) ([]byte, error) {
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

// IsUser checks for consistency in email/pass combination
func isUser(usr *User, password string) error {
	salt, err := base64.StdEncoding.DecodeString(usr.Salt)
	if err != nil {
		return err
	}

	err = checkPassword([]byte(usr.Hash), []byte(password), salt)
	if err != nil {
		return err
	}

	return nil
}

func getUser(email string, repo repository.Repository) (*User, error) {
	userByte, err := repo.User(email)
	if err != nil {
		return nil, err
	}

	if userByte == nil {
		return nil, nil
	}

	user := &User{}
	err = json.Unmarshal(userByte, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// checkPassword compares the hash with the salted password. A nil return means
// the password is correct, but an error could mean either the password is not
// correct, or the salt process failed - indicated in logs
func checkPassword(hash, password, salt []byte) error {
	salted, err := saltPassword(password, salt)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword(hash, salted)
}
