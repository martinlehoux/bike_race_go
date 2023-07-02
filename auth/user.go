package auth

import (
	"bike_race/core"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrBadPassword          = errors.New("incorrect password")
	ErrUserUsernameTooShort = errors.New("username must be at least 3 characters")
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
		return user, ErrUserUsernameTooShort
	}
	user.Id = core.NewID()
	user.Username = username
	user.Language = "en"
	return user, nil
}

func (user *User) SetPassword(oldPassword string, newPassword string) error {
	if user.PasswordHash != nil && bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(oldPassword)) != nil {
		return ErrBadPassword
	}
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return core.Wrap(err, "failed to generate password hash")
	}
	user.PasswordHash = newPasswordHash
	return nil
}
