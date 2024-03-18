package entity

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/admin"
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
	"github.com/gohugonet/hugoverse/internal/domain/admin/valueobject"
)

type Administrator struct {
	Repo repository.Repository
}

func (a *Administrator) ValidateUser(email, password string) error {
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

func (a *Administrator) NewUser(email, password string) (admin.User, error) {
	userData, err := getUser(email, a.Repo)
	if err != nil {
		return nil, err
	}
	if userData != nil {
		return nil, errors.New("user already exists")
	}

	salt, err := valueobject.RandSalt()
	if err != nil {
		return nil, err
	}

	hash, err := valueobject.HashPassword([]byte(password), salt)
	if err != nil {
		return nil, err
	}

	id, err := a.Repo.NextUserId(email)
	if err != nil {
		return nil, err
	}

	user := &valueobject.User{
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

// IsUser checks for consistency in email/pass combination
func isUser(usr *valueobject.User, password string) error {
	salt, err := base64.StdEncoding.DecodeString(usr.Salt)
	if err != nil {
		return err
	}

	err = valueobject.CheckPassword([]byte(usr.Hash), []byte(password), salt)
	if err != nil {
		return err
	}

	return nil
}

func getUser(email string, repo repository.Repository) (*valueobject.User, error) {
	userByte, err := repo.User(email)
	if err != nil {
		return nil, err
	}

	if userByte == nil {
		return nil, nil
	}

	user := &valueobject.User{}
	err = json.Unmarshal(userByte, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
