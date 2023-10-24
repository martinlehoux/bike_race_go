package auth

import (
	"errors"

	"github.com/martinlehoux/kagamigo/kcore"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrBadPassword          = errors.New("incorrect password")
	ErrUserUsernameTooShort = errors.New("username must be at least 3 characters")
)

type User struct {
	Id           kcore.ID
	Username     string
	PasswordHash []byte
	language     string
}

func (user User) Language() string {
	return user.language
}

func NewUser(username string) (User, error) {
	var user User
	if len(username) < 3 {
		return user, ErrUserUsernameTooShort
	}
	user.Id = kcore.NewID()
	user.Username = username
	user.language = "en"
	return user, nil
}

func (user *User) SetPassword(oldPassword string, newPassword string) error {
	if user.PasswordHash != nil && bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(oldPassword)) != nil {
		return ErrBadPassword
	}
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return kcore.Wrap(err, "failed to generate password hash")
	}
	user.PasswordHash = newPasswordHash
	return nil
}
