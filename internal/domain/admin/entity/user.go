package entity

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/mdfriday/hugoverse/internal/domain/admin"
	"github.com/mdfriday/hugoverse/internal/domain/admin/repository"
	"github.com/mdfriday/hugoverse/internal/domain/admin/valueobject"
	"github.com/mdfriday/hugoverse/pkg/loggers"
)

type Administrator struct {
	Repo repository.Repository

	Log loggers.Logger
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

func (a *Administrator) IsUserExists(email string) bool {
	user, err := getUser(email, a.Repo)
	if err != nil {
		a.Log.Errorf("Error [IsUserExists] checking user: %v", err)
		return false
	}
	return user != nil
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

func (a *Administrator) GetUser(email string) (admin.User, error) {
	return getUser(email, a.Repo)
}

func (a *Administrator) AllUsersExcept(one admin.User) ([]admin.User, error) {
	users, err := a.AllUsers()
	if err != nil {
		return nil, err
	}

	var result []admin.User
	for _, user := range users {
		if user.Name() != one.Name() {
			result = append(result, user)
		}
	}
	return result, nil
}

func (a *Administrator) AllUsers() ([]admin.User, error) {
	users := make([]admin.User, 0)

	userBytes := a.Repo.Users()
	for _, userByte := range userBytes {
		if userByte == nil {
			return nil, nil
		}

		user := &valueobject.User{}
		err := json.Unmarshal(userByte, user)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
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
