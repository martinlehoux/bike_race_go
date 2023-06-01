package auth

import (
	"bike_race/core"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id           core.ID
	Username     string
	PasswordHash []byte
	Language     string
}

func NewUser(username string) (User, error) {
	var user User
	if len(username) < 3 {
		return user, errors.New("username must be at least 3 characters")
	}
	user.Id = core.NewID()
	user.Username = username
	user.Language = "en"
	return user, nil
}

func (user *User) SetPassword(oldPassword string, newPassword string) error {
	if user.PasswordHash != nil && bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(oldPassword)) != nil {
		return errors.New("incorrect password")
	}
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = newPasswordHash
	return nil
}
